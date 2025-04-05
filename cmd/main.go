package main

import (
	"log"
	"time"

	"github.com/BigBr41n/echoAuth/db"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/models"
	"github.com/BigBr41n/echoAuth/services"
	"github.com/google/uuid"

	"github.com/BigBr41n/echoAuth/config"
)

func main() {

	if err := config.Init(); err != nil {
		log.Fatal("Error While Loading Env Vars")
	}

	logger.InitLogger(logger.DefaultConfig())

	db.ConnectDB()

	userRepo := models.NewUserRepo(db.DBPool)
	userService := services.NewUserService(userRepo)

	userID, err := userService.SignUp(&models.User{
		ID:        uuid.UUID{},
		Username:  "R4him",
		Email:     "testing@gmaol.co",
		Password:  "pass#--*--$7R0NG",
		Role:      "client",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		logger.Error(err)
	}

	logger.Info(userID)

	logger.Info("Application started")
	logger.WithField("user_id", 123).Info("User logged in")

}
