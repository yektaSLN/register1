package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"login/dto"
	"login/kafka"
	"login/models"
	"login/repository"
	"login/utils"
	"login/validator"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(request dto.RegisterRequest) (*dto.UserResponse, error)
	Login(request dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(request dto.RefreshTokenRequest) (*dto.AuthResponse, error)
	ForgotPassword(request dto.ForgotPasswordRequest) (string, error)
	ResetPassword(request dto.ResetPasswordRequest) error
	GetUserByID(id uint) (*dto.UserResponse, error)
}

type authService struct {
	userRepository repository.UserRepository
	jwtSecret      string
	jwtExpiration  int
	validator      *validator.Validator
	redisClient    *redis.Client
	kafkaProducer  *kafka.Producer
}

func NewAuthService(
	userRepository repository.UserRepository,
	jwtSecret string,
	jwtExpiration int,
	redisClient *redis.Client,
	kafkaProducer *kafka.Producer,
) AuthService {
	return &authService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
		jwtExpiration:  jwtExpiration,
		validator:      validator.New(),
		redisClient:    redisClient,
		kafkaProducer:  kafkaProducer,
	}
}

func (s *authService) Register(
	request dto.RegisterRequest,
) (*dto.UserResponse, error) {

	request.Username = validator.NormalizeUsername(request.Username)
	request.Email = validator.NormalizeEmail(request.Email)
	request.Phone = validator.NormalizePhone(request.Phone)

	if err := s.validator.Validate(request); err != nil {
		return nil, err
	}

	_, err := s.userRepository.FindByUsername(request.Username)
	if err == nil {
		return nil, utils.ErrUsernameExists
	}

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

	eventPayload := map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"age":      user.Age,
		"phone":    user.Phone,
	}

	if err := s.kafkaProducer.Publish(
		context.Background(),
		kafka.EventUserRegistered,
		eventPayload,
	); err != nil {
		log.Printf("Kafka publish failed for user.registered: %v", err)
	}

	return &dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Age:      user.Age,
		Phone:    user.Phone,
	}, nil
}

func (s *authService) Login(
	request dto.LoginRequest,
) (*dto.AuthResponse, error) {

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

	eventPayload := map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
	}

	if err := s.kafkaProducer.Publish(
		context.Background(),
		kafka.EventUserLoggedIn,
		eventPayload,
	); err != nil {
		log.Printf("Kafka publish failed for user.logged_in: %v", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = s.redisClient.Set(
		context.Background(),
		"refresh:"+refreshToken,
		user.ID,
		7*24*time.Hour,
	).Err()

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
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) RefreshToken(
	request dto.RefreshTokenRequest,
) (*dto.AuthResponse, error) {

	userID, err := s.redisClient.Get(
		context.Background(),
		"refresh:"+request.RefreshToken,
	).Uint64()

	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	user, err := s.userRepository.FindByID(uint(userID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrUserNotFound
		}

		return nil, err
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

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *authService) ForgotPassword(
	request dto.ForgotPasswordRequest,
) (string, error) {

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

func (s *authService) ResetPassword(
	request dto.ResetPasswordRequest,
) error {

	if err := s.validator.Validate(request); err != nil {
		return err
	}

	token, err := utils.ValidateToken(
		request.Token,
		s.jwtSecret,
	)

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

	if err := s.userRepository.Update(user); err != nil {
		return err
	}

	eventPayload := map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
	}

	if err := s.kafkaProducer.Publish(
		context.Background(),
		kafka.EventPasswordReset,
		eventPayload,
	); err != nil {
		log.Printf("Kafka publish failed for user.password_reset: %v", err)
	}

	return nil

}

func (s *authService) GetUserByID(
	id uint,
) (*dto.UserResponse, error) {

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
