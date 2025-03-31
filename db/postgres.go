package db

import (
	"context"
	"fmt"
	"time"

	"github.com/BigBr41n/echoAuth/config"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool

func ConnectDB() {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
	)

	var err error
	for i := 0; i < 3; i++ {
		DBPool, err = pgxpool.New(context.Background(), dbURL)
		if err == nil {
			logger.Info("Connected to PostgresSQL")
			break
		}
		logger.Error(fmt.Sprintf("Failed to connect to DB (attempt %d) : %v", i+1, err))
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logger.Fatal(fmt.Sprintf("Databse connection failed : %v ", err))
	}
}

func Close() {
	DBPool.Close()
}
