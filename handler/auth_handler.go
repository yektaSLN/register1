package handler

import (
	"errors"
	"net/http"

	"login/dto"
	"login/service"
	"login/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	//we get the service the main file has built
	authService service.AuthService
}

// constructor
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var request dto.RegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}

	user, err := h.authService.Register(request)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request dto.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}

	response, err := h.authService.Login(request)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// handling  requests for generating a new access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var request dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}

	response, err := h.authService.RefreshToken(request)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var request dto.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}

	token, err := h.authService.ForgotPassword(request)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "copy this token to reset youe password",
		"token":   token,
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var request dto.ResetPasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleError(c, utils.ErrInvalidRequest)
		return
	}

	if err := h.authService.ResetPassword(request); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "password reset successfully",
	})
}

// somehow like the profile panel
func (h *AuthHandler) GetMe(c *gin.Context) {

	//getting the user_id from middleware(the output is an interface!!)
	userIDValue, exists := c.Get("user_id")

	if !exists {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}

	//converting anything in the interface to uint
	userID, ok := userIDValue.(uint)
	if !ok {
		h.handleError(c, utils.ErrInvalidToken)
		return
	}

	//handler->service->repository->database
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user,
	})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	var appErr *utils.AppError

	if errors.As(err, &appErr) {
		c.JSON(
			appErr.StatusCode,
			utils.NewErrorResponse(appErr),
		)
		return
	}

	//if the errors arent in apperr just calaim its internal error
	c.JSON(
		http.StatusInternalServerError,
		utils.NewErrorResponse(utils.ErrInternalServer),
	)
}
