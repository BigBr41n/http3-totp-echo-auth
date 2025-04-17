package routes

import (
	"github.com/BigBr41n/echoAuth/controllers"
	ctm "github.com/BigBr41n/echoAuth/internal/custom_middlewares"
	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(api *echo.Group, authCtl controllers.AuthControllerI) {
	userRoute := api.Group("/auth")

	userRoute.POST("/signup", authCtl.RegisterNewUser)
	userRoute.POST("/login", authCtl.LoginUser)
	userRoute.POST("/refresh", authCtl.RefreshAxsToken)
	userRoute.POST("/2FA/enable", authCtl.Enable2FA, ctm.JwtAuthMidd)
	userRoute.POST("/validate-totp", authCtl.ValidateTOTP)
}
