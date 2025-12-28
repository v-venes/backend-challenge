package repository

import (
	"github.com/v-venes/backend-challenge/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockRepository interface {
	UpsertBatch(stocks []domain.Stock) error
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
				{Name: "trade_date"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"price",
				"quantity",
			}),
		}).
		CreateInBatches(stocks, len(stocks)).Error
}
