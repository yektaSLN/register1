package service

import (
	"errors"

	"login/dto"
	"login/models"
	"login/repository"
	"login/utils"
	"login/validator"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(request dto.RegisterRequest) (*dto.UserResponse, error)
	Login(request dto.LoginRequest) (*dto.AuthResponse, error)
	ForgotPassword(request dto.ForgotPasswordRequest) (string, error)
	ResetPassword(request dto.ResetPasswordRequest) error
	GetUserByID(id uint) (*dto.UserResponse, error)
}

type authService struct {
	userRepository repository.UserRepository
	jwtSecret      string
	jwtExpiration  int
	validator      *validator.Validator
}

// constructor
func NewAuthService(
	userRepository repository.UserRepository,
	jwtSecret string,
	jwtExpiration int,
) AuthService {
	return &authService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
		jwtExpiration:  jwtExpiration,
		validator:      validator.New(),
	}
}

func (s *authService) Register(request dto.RegisterRequest) (*dto.UserResponse, error) {

	request.Username = validator.NormalizeUsername(request.Username)
	request.Email = validator.NormalizeEmail(request.Email)
	request.Phone = validator.NormalizePhone(request.Phone)

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	// checking if the username exists in database
	_, err := s.userRepository.FindByUsername(request.Username)
	if err == nil {
		return nil, utils.ErrUsernameExists
	}

	// the username doesn't exist so it's ok
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	_, err = s.userRepository.FindByEmail(request.Email)
	if err == nil {
		return nil, utils.ErrEmailExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	_, err = s.userRepository.FindByPhone(request.Phone)
	if err == nil {
		return nil, utils.ErrPhoneExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	// saving data in database
	user := &models.User{
		Username: request.Username,
		Email:    request.Email,
		Age:      request.Age,
		Phone:    request.Phone,
		Password: hashedPassword,
	}

	if err := s.userRepository.Create(user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Age:      user.Age,
		Phone:    user.Phone,
	}, nil
}

func (s *authService) Login(request dto.LoginRequest) (*dto.AuthResponse, error) {

	request.Username = validator.NormalizeUsername(request.Username)

	user, err := s.userRepository.FindByUsername(request.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidCredentials
		}

		return nil, err
	}

	if !utils.CheckPassword(user.Password, request.Password) {
		return nil, utils.ErrInvalidCredentials
	}

	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		s.jwtSecret,
		s.jwtExpiration,
	)

	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User: dto.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Age:      user.Age,
			Phone:    user.Phone,
		},
		Token: token,
	}, nil
}

func (s *authService) ForgotPassword(request dto.ForgotPasswordRequest) (string, error) {

	request.Email = validator.NormalizeEmail(request.Email)

	user, err := s.userRepository.FindByEmail(request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", utils.ErrUserNotFound
		}

		return "", err
	}

	token, err := utils.GenerateToken(
		user.ID,
		user.Username,
		s.jwtSecret,
		1,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *authService) ResetPassword(request dto.ResetPasswordRequest) error {

	if err := s.validator.Validate(request); err != nil {
		return err
	}

	// validating token
	token, err := utils.ValidateToken(request.Token, s.jwtSecret)

	if err != nil || !token.Valid {
		return utils.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return utils.ErrInvalidToken
	}

	userIDValue, ok := claims["user_id"]
	if !ok {
		return utils.ErrInvalidToken
	}

	userID, ok := userIDValue.(float64)
	if !ok {
		return utils.ErrInvalidToken
	}

	user, err := s.userRepository.FindByID(uint(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrUserNotFound
		}

		return err
	}

	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword

	return s.userRepository.Update(user)
}

func (s *authService) GetUserByID(id uint) (*dto.UserResponse, error) {

	user, err := s.userRepository.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrUserNotFound
		}

		return nil, err
	}

	return &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Age:      user.Age,
		Phone:    user.Phone,
	}, nil
}
