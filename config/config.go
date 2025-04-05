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
	DBHost     string
	DBPort     string
	ServerPort string
	ENV        string
	JWTSEC     string
	JWTREFSEC  string
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
			DBHost:     os.Getenv("DB_HOST"),
			DBPort:     os.Getenv("DB_PORT"),
			ServerPort: os.Getenv("SERVER_PORT"),
			ENV:        os.Getenv("ECHO_AUTH_APP"),
			JWTSEC:     os.Getenv("JWT_SECRET"),
			JWTREFSEC:  os.Getenv("JWT-REF-SEC"),
		}

		log.Println("Configuration loaded successfully")

		return nil
	}

	return err

}
