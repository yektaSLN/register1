package validator

import (
	"regexp"
	"strings"

	"login/utils"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		match, _ := regexp.MatchString(`^\+?[0-9]{10,15}$`, fl.Field().String())
		return match
	})

	_ = v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		if len(password) < 8 {
			return false
		}

		hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
		hasLower, _ := regexp.MatchString(`[a-z]`, password)
		hasNumber, _ := regexp.MatchString(`[0-9]`, password)

		return hasUpper && hasLower && hasNumber
	})

	return &Validator{
		validate: v,
	}
}

func (v *Validator) Validate(data interface{}) error {
	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return utils.NewValidationError("invalid request")
	}

	firstError := validationErrors[0]

	switch firstError.Field() {
	case "Username":
		return utils.NewValidationError("invalid username")

	case "Email":
		return utils.NewValidationError("invalid email")

	case "Age":
		return utils.NewValidationError("invalid age")

	case "Phone":
		return utils.NewValidationError("invalid phone")

	case "Password", "NewPassword":
		return utils.NewValidationError(
			"password must contain at least 8 characters, one uppercase letter, one lowercase letter and one number",
		)

	case "Token":
		return utils.NewValidationError("invalid token")
	}

	return utils.NewValidationError("invalid request")
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func NormalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}
