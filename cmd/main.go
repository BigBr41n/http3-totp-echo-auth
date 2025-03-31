package main

import (
	"log"

	"github.com/BigBr41n/echoAuth/internal/logger"

	"github.com/BigBr41n/echoAuth/config"
)

func main() {

	if err := config.Init(); err != nil {
		log.Fatal("Error While Loading Env Vars")
	}

	logger.InitLogger(logger.DefaultConfig())

	logger.Info("Application started")
	logger.WithField("user_id", 123).Info("User logged in")

}
