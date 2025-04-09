package custommiddlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		ip := c.Request().Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = c.Request().RemoteAddr
		}

		err := next(c)

		duration := time.Since(start)

		zap.L().Info("request",
			zap.String("method", c.Request().Method),
			zap.String("path", c.Request().URL.Path),
			zap.String("ip", ip),
			zap.Int("status", c.Response().Status),
			zap.Duration("duration", duration),
		)

		return err
	}
}
