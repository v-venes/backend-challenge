package service

import (
	"errors"
	"log"
	"time"

	"github.com/v-venes/backend-challenge/internal/domain"
	"github.com/v-venes/backend-challenge/internal/repository"
)

type StockService struct {
	repo repository.StockRepository
}

func NewStockService(repo repository.StockRepository) *StockService {
	return &StockService{repo: repo}
}

func (s *StockService) GetAggregated(ticker string, startDateStr string) (*domain.AggregatedStockResult, error) {

	if ticker == "" {
		return nil, errors.New("ticker is required")
	}

	loc, _ := time.LoadLocation("America/Sao_Paulo")
	today := time.Now().In(loc).Truncate(24 * time.Hour)
	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		endDate = today
		startDate = today.AddDate(0, 0, -7)
	} else {
		startDate, err = time.ParseInLocation("2006-01-02", startDateStr, loc)
		if err != nil {
			return nil, errors.New("invalid start_date format, expected YYYY-MM-DD")
		}
		endDate = today
	}

	log.Printf("[INFO] fetching aggregated data for ticker %s since %s - %s", ticker, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	aggregatedTicker, err := s.repo.GetAggregatedByTicker(ticker, startDate.UTC(), endDate.UTC())

	if err != nil {
		return nil, err
	}

	return aggregatedTicker, err
}
