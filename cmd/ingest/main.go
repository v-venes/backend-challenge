package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/v-venes/backend-challenge/internal/config"
	"github.com/v-venes/backend-challenge/internal/ingest"
	"github.com/v-venes/backend-challenge/internal/repository"
)

const DATA_PATH = "data/"

func init() {
	if os.Getenv("GO_ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("[ERROR] %s", err.Error())
		}
	}
}

func main() {
	cfg := config.Load()
	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		DSN: fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBUser,
			cfg.DBPass,
			cfg.DBName,
			cfg.DBPort,
		),
	})

	if err != nil {
		log.Fatalf("[ERROR] %s", err.Error())
	}

	repo := repository.NewStockRepository(db)

	ingestService := ingest.NewIngestService(ingest.NewIngestServiceParams{
		BatchSize: 2000,
		Workers:   runtime.NumCPU() * 2,
		Repo:      repo,
	})

	ingestService.Run(DATA_PATH)
}
