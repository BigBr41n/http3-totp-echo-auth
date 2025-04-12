package custommiddlewares

import (
	"net/http"
	"os"
	"runtime/debug"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func RecoverWithJSON() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					// Log the panic and stacktrace
					zap.L().Error("PANIC recovered",
						zap.Any("panic", r),
						zap.String("stack", string(debug.Stack())))

					errMsg := "Internal Server Error"
					if os.Getenv("ECHO_AUTH_APP") != "prod" {
						errMsg = r.(string)
					}

					_ = c.JSON(http.StatusInternalServerError, dtos.ApiErr{
						Status:  http.StatusInternalServerError,
						Err:     errMsg,
						Code:    "INTERNAL_ERROR",
						Details: nil,
					})
				}
			}()

			return next(c)
		}
	}
}
