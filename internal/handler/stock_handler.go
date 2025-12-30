package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/v-venes/backend-challenge/internal/service"
	"gorm.io/gorm"
)

type StockHandler struct {
	service *service.StockService
}

func NewStockHandler(service *service.StockService) *StockHandler {
	return &StockHandler{service: service}
}

func (s *StockHandler) GetAggregated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	ticker := q.Get("ticker")
	startDate := q.Get("data_inicio")

	result, err := s.service.GetAggregated(ticker, startDate)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "ticker not found", http.StatusNotFound)
			return
		}
		log.Printf("[ERROR] %s", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
