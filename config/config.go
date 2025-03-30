package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

var AppConfig Config

func Init() error {

	// check for prodution env
	if os.Getenv("ECHO_AUTH_APP") != "prod" {
		// use instead .env
		if err := godotenv.Load(); err != nil {
			return err
		}
	}

	AppConfig = Config{
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}

	log.Println("Configuration loaded successfully")
	return nil
}
