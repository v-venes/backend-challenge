package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/v-venes/backend-challenge/internal/config"
	"github.com/v-venes/backend-challenge/internal/domain"
	"github.com/v-venes/backend-challenge/internal/repository"
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

	if !db.Migrator().HasTable(&domain.Stock{}) {
		log.Println("[INFO] creating stocks table...")
		err = db.Migrator().CreateTable(&domain.Stock{})
		if err != nil {
			log.Fatalf("[ERROR] %s", err.Error())
		}

		log.Println("[INFO] creating upsert indexes...")
		if err := db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_stocks_ticker_trade_at
			ON stocks (ticker, trade_at);
		`).Error; err != nil {
			log.Fatalf("[ERROR] %s", err.Error())
		}

		log.Println("[INFO] creating query indexes...")
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_stocks_ticker_trade_date
			ON stocks (ticker, trade_date)
			INCLUDE (price, quantity);
		`).Error; err != nil {
			log.Fatalf("[ERROR] %s", err.Error())
		}
	}

	log.Println("[INFO] migrations applied successfully")

}
