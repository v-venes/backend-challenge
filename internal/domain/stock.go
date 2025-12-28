package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DATE_LAYOUT = "2006-01-02 150405000"

type Stock struct {
	ID       uint      `gorm:"primaryKey"`
	Ticker   string    `gorm:"size:20;not null"`
	TradeAt  time.Time `gorm:"not null"`
	Price    float64   `gorm:"not null"`
	Quantity int64     `gorm:"not null"`
}

func (Stock) TableName() string {
	return "stocks"
}

func StockFromStrArray(values map[string]string) (*Stock, error) {

	dateTimeStr := fmt.Sprintf("%s %s", values["datanegocio"], values["horafechamento"])

	tradeAt, err := time.ParseInLocation(DATE_LAYOUT, dateTimeStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("invalid tradeAt: %w", err)
	}

	price, err := strconv.ParseFloat(values["preconegocio"], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	quantity, err := strconv.ParseInt(values["quantidadenegociada"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	return &Stock{
		TradeAt:  tradeAt,
		Ticker:   strings.ToUpper(values["codigoinstrumento"]),
		Price:    price,
		Quantity: quantity,
	}, nil
}
