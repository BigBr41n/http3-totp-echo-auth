package routes

import (
	"github.com/BigBr41n/echoAuth/controllers"
	"github.com/labstack/echo/v4"
)

func RegisterUserRoutes(api *echo.Group, authCtl controllers.AuthControllerI) {
	userRoute := api.Group("/auth")

	userRoute.POST("/signup", authCtl.RegisterNewUser)
	userRoute.POST("/login", authCtl.RegisterNewUser)
}
