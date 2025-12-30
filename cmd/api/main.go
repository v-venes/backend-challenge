package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/v-venes/backend-challenge/internal/config"
	"github.com/v-venes/backend-challenge/internal/handler"
	"github.com/v-venes/backend-challenge/internal/repository"
	"github.com/v-venes/backend-challenge/internal/service"
)

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
	service := service.NewStockService(repo)
	handler := handler.NewStockHandler(service)

	http.HandleFunc("/stocks/aggregate", handler.GetAggregated)

	log.Println("[INFO] API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
