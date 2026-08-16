package utils

import "net/http"

type AppError struct {
	Code       int
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

//read the errors.md file

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       1010,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

var (
	ErrUserNotFound = &AppError{
		Code:       1001,
		Message:    "user not found",
		StatusCode: http.StatusNotFound,
	}

	ErrProductNotFound = &AppError{
		Code:       1002,
		Message:    "product not found",
		StatusCode: http.StatusNotFound,
	}

	ErrInvalidCredentials = &AppError{
		Code:       1003,
		Message:    "invalid username or password",
		StatusCode: http.StatusUnauthorized,
	}

	ErrUsernameExists = &AppError{
		Code:       1004,
		Message:    "username already exists",
		StatusCode: http.StatusConflict,
	}

	ErrEmailExists = &AppError{
		Code:       1005,
		Message:    "email already exists",
		StatusCode: http.StatusConflict,
	}

	ErrPhoneExists = &AppError{
		Code:       1006,
		Message:    "phone already exists",
		StatusCode: http.StatusConflict,
	}

	ErrInvalidToken = &AppError{
		Code:       1007,
		Message:    "invalid or expired token",
		StatusCode: http.StatusUnauthorized,
	}

	ErrInternalServer = &AppError{
		Code:       1008,
		Message:    "internal server error",
		StatusCode: http.StatusInternalServerError,
	}

	ErrInvalidRequest = &AppError{
		Code:       1009,
		Message:    "invalid request body",
		StatusCode: http.StatusBadRequest,
	}
)

type ErrorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
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
