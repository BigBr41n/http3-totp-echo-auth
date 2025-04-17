package custommiddlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ResponseHeadersMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Security Headers
		c.Response().Header().Set("X-Content-Type-Options", "nosniff")
		c.Response().Header().Set("X-Frame-Options", "DENY")
		c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
		c.Response().Header().Set("Referrer-Policy", "no-referrer")
		c.Response().Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Response().Header().Set("Cache-Control", "no-store")
		c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")
		c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=()")

		c.Response().Header().Set("Access-Control-Allow-Origin", "*") //currently no domains
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Response().Header().Get("Content-Type") == "" {
			c.Response().Header().Set("Content-Type", "application/json")
		}

		if c.Request().Method == http.MethodOptions {
			return c.NoContent(http.StatusOK)
		}

		return next(c)
	}
}
