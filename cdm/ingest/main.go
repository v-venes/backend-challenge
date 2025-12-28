package main

import (
	"log"
	"runtime"

	"github.com/v-venes/backend-challenge/internal/ingest"
	"github.com/v-venes/backend-challenge/internal/repository"
)

const ASSETS_PATH = "assets/"

func main() {
	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		DSN: "",
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

	ingestService.Run(ASSETS_PATH)
}
