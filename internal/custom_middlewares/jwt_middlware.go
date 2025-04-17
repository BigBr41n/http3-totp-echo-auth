package custommiddlewares

import (
	"net/http"
	"os"
	"strings"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/utils/jwtImpl"
	"github.com/BigBr41n/echoAuth/utils/response"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func JwtAuthMidd(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.ErrResp(c, &dtos.ApiErr{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_ACCESS_TOKEN",
				Err:     "Missing or invalid Authorization header",
				Details: nil,
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, val, err := jwtImpl.ParseExtractClaims(tokenStr, "access", string(jwtSecret))
		if err != nil {
			return response.ErrResp(c, &dtos.ApiErr{
				Status:  http.StatusInternalServerError,
				Code:    "INTERNAL_ERROR",
				Err:     "Something went wrong, try later",
				Details: nil,
			})
		}
		if !val {
			return response.ErrResp(c, &dtos.ApiErr{
				Status:  http.StatusUnauthorized,
				Code:    "EXPIRED_TOKEN",
				Err:     "Access token is expired use refresh token",
				Details: nil,
			})
		}

		if claims, ok := token.Claims.(*jwtImpl.CustomAccessTokenClaims); ok {
			c.Set("User", claims)
		} else {
			return response.ErrResp(c, &dtos.ApiErr{
				Status:  http.StatusUnauthorized,
				Code:    "INVALID_CLAIMS",
				Err:     "Invalid token claims",
				Details: nil,
			})
		}

		return next(c)
	}
}
