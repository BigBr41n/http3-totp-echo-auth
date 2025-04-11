package response

import (
	"os"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/labstack/echo/v4"
)

// Success Response
func ValResp(c echo.Context, status int, code string, message string, data interface{}) error {
	return c.JSON(status, dtos.ValidResponse{
		Status:  status,
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Error Response
func ErrResp(c echo.Context, status int, code string, errMsg string, details interface{}) error {
	if (os.Getenv("ECHO_AUTH_APP") == "prod") && (status == 500) {
		errMsg = "Internal Server Error"
	}
	return c.JSON(status, dtos.ErrResponse{
		Status:  status,
		Code:    code,
		Error:   errMsg,
		Details: details,
	})
}
