package validator

import (
	"fmt"

	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/go-playground/validator/v10"
)

// geniric function to validate user inputs
var validate = validator.New()

func Validate(data interface{}) error {
	err := validate.Struct(data)
	if err != nil {
		logger.Error("validation failed", err)
		return fmt.Errorf("validation failed : %v", err)
	}
	return nil
}
