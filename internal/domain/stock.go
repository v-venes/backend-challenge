package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DATE_LAYOUT = "2006-01-02 150405.000"

type Stock struct {
	ID        uint      `gorm:"primaryKey"`
	Ticker    string    `gorm:"type:varchar(16);not null"`
	TradeAt   time.Time `gorm:"type:timestamptz;not null"`
	TradeDate time.Time `gorm:"type:date;not null"`
	Price     float64   `gorm:"type:numeric(15,3);not null"`
	Quantity  int64     `gorm:"not null"`
}

type AggregatedStockResult struct {
	Ticker         string  `json:"ticker"`
	MaxRangeValue  float64 `json:"max_range_value"`
	MaxDailyVolume int64   `json:"max_daily_volume"`
}

func (Stock) TableName() string {
	return "stocks"
}

func StockFromStrArray(values map[string]string) (*Stock, error) {
	date := values["datanegocio"]
	timeRaw := values["horafechamento"]

	if len(timeRaw) != 9 {
		return nil, fmt.Errorf("invalid horafechamento length: %s", timeRaw)
	}

	timeFormatted := fmt.Sprintf(
		"%s.%s",
		timeRaw[:6],
		timeRaw[6:],
	)

	dateTimeStr := fmt.Sprintf("%s %s", date, timeFormatted)

	tradeAt, err := time.ParseInLocation(DATE_LAYOUT, dateTimeStr, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("invalid tradeAt: %w", err)
	}

	price, err := parseDecimalBR(values["preconegocio"])
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	quantity, err := strconv.ParseInt(values["quantidadenegociada"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	tradeDate := tradeAt.Truncate(24 * time.Hour)

	return &Stock{
		TradeAt:   tradeAt,
		TradeDate: tradeDate,
		Ticker:    strings.ToUpper(values["codigoinstrumento"]),
		Price:     price,
		Quantity:  quantity,
	}, nil
}

func parseDecimalBR(value string) (float64, error) {
	normalized := strings.Replace(value, ",", ".", 1)
	return strconv.ParseFloat(normalized, 64)
}
