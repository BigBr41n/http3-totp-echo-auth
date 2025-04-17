package main

import (
	"crypto/tls"
	"log"
	"net/http"

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
	"github.com/quic-go/quic-go/http3"
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
	e.Use(cstm_mdlwr.ResponseHeadersMiddleware)

	// register global custom group
	api := e.Group("/api/v1")

	// register /user routes
	routes.RegisterUserRoutes(api, authControllers)

	// http 3 setup
	tlsCert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h3", "h2", "http/1.1"},
	}

	handler := e.Server.Handler

	// Create HTTP/3 server
	h3Server := &http3.Server{
		Addr:      ":8443",
		TLSConfig: tlsConf,
		Handler:   handler,
	}

	// Run HTTP/1.1 + HTTP/2 (TCP)
	go func() {
		logger.Info("HTTP/1.1 and HTTP/2 running on TCP :8443...")
		httpServer := &http.Server{
			Addr:      ":8443",
			TLSConfig: tlsConf,
			Handler:   handler,
		}
		if err := httpServer.ListenAndServeTLS("server.crt", "server.key"); err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	// start the http3 server
	if err := h3Server.ListenAndServe(); err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
