package main

import (
	"log"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/config"
	"github.com/BigBr41n/echoAuth/db"
	"github.com/BigBr41n/echoAuth/db/sqlc"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/services"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {

	if err := config.Init(); err != nil {
		log.Fatal("Error While Loading Env Vars")
	}

	logger.InitLogger(logger.DefaultConfig())

	db.ConnectDB()

	queries := sqlc.New(db.DBPool)

	userService := services.NewUserService(queries)

	var pgUUID pgtype.UUID
	pgUUID.Bytes = uuid.New()

	userID, err := userService.SignUp(&dtos.CreateUserDTO{
		Username: "R4him",
		Email:    "testing@gmaol.co",
		Password: "pass#--*--$7R0NG",
		Role:     "client",
	})

	if err != nil {
		logger.Error(err)
	}

	logger.Info(userID)

	logger.Info("Application started")
	logger.WithField("user_id", 123).Info("User logged in")

}
