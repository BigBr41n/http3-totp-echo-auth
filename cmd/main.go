package main

import (
	"log"

	"github.com/BigBr41n/echoAuth/config"
	"github.com/BigBr41n/echoAuth/controllers"
	"github.com/BigBr41n/echoAuth/db"
	"github.com/BigBr41n/echoAuth/db/sqlc"
	cstm_mdlwr "github.com/BigBr41n/echoAuth/internal/custom_middlewares"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/routes"
	"github.com/BigBr41n/echoAuth/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	// load env vars if in dev env
	if err := config.Init(); err != nil {
		log.Fatal("Error While Loading Env Vars")
	}

	// init the logger
	logger.Init()

	// coonnect to DB
	db.ConnectDB()

	// init SQLC queries
	queries := sqlc.New(db.DBPool)

	// creating auth service and controller
	authService := services.NewAuthService(queries)
	authControllers := controllers.NewAuthController(authService)

	// echo instance & middlewares
	e := echo.New()
	e.Use(cstm_mdlwr.LoggerMiddleware)
	//e.Use(middleware.Recover())
	e.Use(cstm_mdlwr.RecoverWithJSON())
	e.Use(middleware.CORS())

	// register global custom group
	api := e.Group("/api/v1")

	// register /user routes
	routes.RegisterUserRoutes(api, authControllers)

	// start the server
	if err := e.Start(":8080"); err != nil {
		e.Logger.Fatal("Server failed to start: ", err)
	}
}
