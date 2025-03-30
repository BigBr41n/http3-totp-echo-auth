package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

var AppConfig Config

func load() error {
	// check for prodution env
	if os.Getenv("ECHO_AUTH_APP") != "prod" {
		// use instead .env
		if err := godotenv.Load(); err != nil {
			return err
		}
	}
	return nil
}

func Init() error {

	var err error
	retries := 3
	delay := 2 * time.Second

	for i := 0; i < retries; i++ {
		if err = load(); err != nil {
			log.Printf("⚠️ Failed to load .env (attempt %d/%d): %v", i+1, retries, err)
			time.Sleep(delay)
			continue
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

	return err

}
