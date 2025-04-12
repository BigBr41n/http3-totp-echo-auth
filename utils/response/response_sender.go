package response

import (
	"net/http"
	"os"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Success Response
func ValResp(c echo.Context, resp *dtos.ValidResponse) error {
	return c.JSON(resp.Status, dtos.ValidResponse{
		Status:  resp.Status,
		Code:    resp.Code,
		Message: resp.Message,
		Data:    resp.Data,
	})
}

// Error Response
func ErrResp(c echo.Context, resp error) error {
	var ApiError *dtos.ApiErr
	var ok bool
	// assert back the error if possible
	if ApiError, ok = resp.(*dtos.ApiErr); !ok {
		logger.Error("Unknown error", zap.Error(resp))

		return c.JSON(http.StatusInternalServerError, dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Err:     resp.Error(),
			Code:    "INTERNAL_ERROR",
			Details: "Unkown error",
		})
	}

	if (os.Getenv("ECHO_AUTH_APP") == "prod") && (ApiError.Status == 500) {
		ApiError.Err = "Internal Server Error"
	}

	return c.JSON(ApiError.Status, dtos.ApiErr{
		Status:  ApiError.Status,
		Code:    ApiError.Code,
		Err:     ApiError.Err,
		Details: ApiError.Details,
	})
}
