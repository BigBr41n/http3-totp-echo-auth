package validator

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

func isValidPassword(password string) bool {
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+]`).MatchString(password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}

// singleton validator instance
var validate *validator.Validate

// initialized only once
func init() {
	validate = validator.New()

	validate.RegisterValidation("strongpwd", func(fl validator.FieldLevel) bool {
		return isValidPassword(fl.Field().String())
	})

	// Register alias for password validation
	validate.RegisterAlias("pwd", "min=8,max=20,strongpwd")
}

func GetValidator() *validator.Validate {
	return validate
}

func Validate(data interface{}) error {
	err := GetValidator().Struct(data)
	if err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}
	return nil
}
