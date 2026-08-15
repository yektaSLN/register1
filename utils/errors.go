package utils

import "net/http"

type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

var (
	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "user not found",
		StatusCode: http.StatusNotFound,
	}

	ErrProductNotFound = &AppError{
		Code:       "PRODUCT_NOT_FOUND",
		Message:    "product not found",
		StatusCode: http.StatusNotFound,
	}

	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "invalid username or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrUsernameExists = &AppError{
		Code:       "USERNAME_ALREADY_EXISTS",
		Message:    "username already exists",
		StatusCode: http.StatusConflict,
	}

	ErrEmailExists = &AppError{
		Code:       "EMAIL_ALREADY_EXISTS",
		Message:    "email already exists",
		StatusCode: http.StatusConflict,
	}

	ErrPhoneExists = &AppError{
		Code:       "PHONE_ALREADY_EXISTS",
		Message:    "phone already exists",
		StatusCode: http.StatusConflict,
	}

	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "invalid or expired token",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrInvalidRequest = &AppError{
		Code:       "INVALID_REQUEST",
		Message:    "invalid request body",
		StatusCode: http.StatusBadRequest,
	}
)

type ErrorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewErrorResponse(err *AppError) ErrorResponse {
	response := ErrorResponse{
		Success: false,
	}

	response.Error.Code = err.Code
	response.Error.Message = err.Message

	return response
}
