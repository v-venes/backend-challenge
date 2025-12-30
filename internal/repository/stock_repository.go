package repository

import (
	"time"

	"github.com/v-venes/backend-challenge/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockRepository interface {
	UpsertBatch(stocks []domain.Stock) error
	GetAggregatedByTicker(ticker string, startDate time.Time, endDate time.Time) (*domain.AggregatedStockResult, error)
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

func (s *stockRepository) UpsertBatch(stocks []domain.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	return s.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "ticker"},
				{Name: "trade_at"},
			},
			DoNothing: true,
		}).
		CreateInBatches(stocks, len(stocks)).Error
}

func (s *stockRepository) GetAggregatedByTicker(ticker string, startDate time.Time, endDate time.Time) (*domain.AggregatedStockResult, error) {
	var result domain.AggregatedStockResult
	tx := s.db.Raw(`
		SELECT
			? AS ticker,
			MAX(price) AS max_range_value,
			MAX(daily_vol) AS max_daily_volume
		FROM (
			SELECT 
				MAX(price) AS price,
				SUM(quantity) AS daily_vol
			FROM stocks
			WHERE ticker = ?
				AND trade_date >= ?
				AND trade_date < ?
			GROUP BY trade_date
		) daily_stats;
	`, ticker, ticker, startDate, endDate).Scan(&result)
	if tx.Error != nil {
		return nil, tx.Error
	}

	if result.Ticker == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &result, nil
}
