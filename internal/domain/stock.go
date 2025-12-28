package domain

import "time"

type Stock struct {
	TradeDate time.Time
	Ticker    string
	Price     float64
	Quantity  int64
	TradeTime time.Time
}

func StockFromStrArray(values map[string]string) Stock {
	return Stock{}
}
