package ingest

import (
	"testing"
	"time"

	"github.com/v-venes/backend-challenge/internal/domain"
)

func TestNewBatcher(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int
	}{
		{
			name:      "cria batcher com tamanho 10",
			batchSize: 10,
		},
		{
			name:      "cria batcher com tamanho 100",
			batchSize: 100,
		},
		{
			name:      "cria batcher com tamanho 1",
			batchSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batcher := NewBatcher(tt.batchSize)

			if batcher == nil {
				t.Fatal("NewBatcher retornou nil")
			}

			if batcher.size != tt.batchSize {
				t.Errorf("size esperado %d, recebido %d", tt.batchSize, batcher.size)
			}

			if batcher.batch == nil {
				t.Error("batch slice não foi inicializado")
			}

			if len(batcher.batch) != 0 {
				t.Errorf("batch deveria estar vazio, mas tem %d elementos", len(batcher.batch))
			}

			if cap(batcher.batch) != tt.batchSize {
				t.Errorf("capacidade esperada %d, recebida %d", tt.batchSize, cap(batcher.batch))
			}
		})
	}
}

func TestBatcher_Run_CompleteBatch(t *testing.T) {
	batchSize := 3
	batcher := NewBatcher(batchSize)

	stocksCh := make(chan domain.Stock)
	batchesCh := make(chan []domain.Stock)

	stocks := []domain.Stock{
		{Ticker: "PETR4", Quantity: 100},
		{Ticker: "VALE3", Quantity: 200},
		{Ticker: "ITUB4", Quantity: 300},
	}

	go batcher.Run(stocksCh, batchesCh)

	go func() {
		for _, stock := range stocks {
			stocksCh <- stock
		}
		close(stocksCh)
	}()

	batch := <-batchesCh

	if len(batch) != batchSize {
		t.Errorf("esperado batch com %d elementos, recebido %d", batchSize, len(batch))
	}

	for i, stock := range stocks {
		if batch[i].Ticker != stock.Ticker {
			t.Errorf("ticker no índice %d: esperado %s, recebido %s", i, stock.Ticker, batch[i].Ticker)
		}
		if batch[i].Quantity != stock.Quantity {
			t.Errorf("quantity no índice %d: esperado %d, recebido %d", i, stock.Quantity, batch[i].Quantity)
		}
	}

	select {
	case _, ok := <-batchesCh:
		if ok {
			t.Error("canal batchesCh deveria estar fechado")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout esperando canal fechar")
	}
}

func TestBatcher_Run_PartialBatch(t *testing.T) {
	batchSize := 5
	batcher := NewBatcher(batchSize)

	stocksCh := make(chan domain.Stock)
	batchesCh := make(chan []domain.Stock)

	stocks := []domain.Stock{
		{Ticker: "PETR4", Quantity: 100},
		{Ticker: "VALE3", Quantity: 200},
	}

	go batcher.Run(stocksCh, batchesCh)

	go func() {
		for _, stock := range stocks {
			stocksCh <- stock
		}
		close(stocksCh)
	}()

	batch := <-batchesCh

	if len(batch) != len(stocks) {
		t.Errorf("esperado batch com %d elementos, recebido %d", len(stocks), len(batch))
	}

	for i, stock := range stocks {
		if batch[i].Ticker != stock.Ticker {
			t.Errorf("ticker no índice %d: esperado %s, recebido %s", i, stock.Ticker, batch[i].Ticker)
		}
	}
}

func TestBatcher_Run_MultipleBatches(t *testing.T) {
	batchSize := 2
	batcher := NewBatcher(batchSize)

	stocksCh := make(chan domain.Stock)
	batchesCh := make(chan []domain.Stock)

	stocks := []domain.Stock{
		{Ticker: "PETR4", Quantity: 100},
		{Ticker: "VALE3", Quantity: 200},
		{Ticker: "ITUB4", Quantity: 300},
		{Ticker: "BBDC4", Quantity: 400},
		{Ticker: "ABEV3", Quantity: 500},
	}

	go batcher.Run(stocksCh, batchesCh)

	go func() {
		for _, stock := range stocks {
			stocksCh <- stock
		}
		close(stocksCh)
	}()

	batch1 := <-batchesCh
	if len(batch1) != batchSize {
		t.Errorf("batch 1: esperado %d elementos, recebido %d", batchSize, len(batch1))
	}

	batch2 := <-batchesCh
	if len(batch2) != batchSize {
		t.Errorf("batch 2: esperado %d elementos, recebido %d", batchSize, len(batch2))
	}

	batch3 := <-batchesCh
	if len(batch3) != 1 {
		t.Errorf("batch 3: esperado 1 elemento, recebido %d", len(batch3))
	}

	totalProcessed := len(batch1) + len(batch2) + len(batch3)
	if totalProcessed != len(stocks) {
		t.Errorf("esperado processar %d stocks, processou %d", len(stocks), totalProcessed)
	}
}

func TestBatcher_Run_EmptyChannel(t *testing.T) {
	batchSize := 10
	batcher := NewBatcher(batchSize)

	stocksCh := make(chan domain.Stock)
	batchesCh := make(chan []domain.Stock)

	go batcher.Run(stocksCh, batchesCh)

	close(stocksCh)

	select {
	case _, ok := <-batchesCh:
		if ok {
			t.Error("não deveria receber batch de canal vazio")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout esperando canal fechar")
	}
}

func TestBatcher_Run_LargeBatch(t *testing.T) {
	batchSize := 1000
	batcher := NewBatcher(batchSize)

	stocksCh := make(chan domain.Stock, batchSize)
	batchesCh := make(chan []domain.Stock)

	var stocks []domain.Stock
	for i := 0; i < batchSize; i++ {
		stocks = append(stocks, domain.Stock{
			Ticker:   "TEST",
			Quantity: int64(i),
		})
	}

	go batcher.Run(stocksCh, batchesCh)

	go func() {
		for _, stock := range stocks {
			stocksCh <- stock
		}
		close(stocksCh)
	}()

	batch := <-batchesCh

	if len(batch) != batchSize {
		t.Errorf("esperado batch com %d elementos, recebido %d", batchSize, len(batch))
	}

	for i := 0; i < batchSize; i++ {
		if batch[i].Quantity != int64(i) {
			t.Errorf("índice %d: esperado quantity %d, recebido %d", i, i, batch[i].Quantity)
		}
	}
}

func TestBatcher_Run_ConcurrentSafety(t *testing.T) {
	batchSize := 10
	batcher := NewBatcher(batchSize)

	stocksCh := make(chan domain.Stock, 100)
	batchesCh := make(chan []domain.Stock, 10)

	numStocks := 95

	go batcher.Run(stocksCh, batchesCh)

	go func() {
		for i := 0; i < numStocks; i++ {
			stocksCh <- domain.Stock{
				Ticker:   "CONCURRENT",
				Quantity: int64(i),
			}
		}
		close(stocksCh)
	}()

	var totalReceived int
	var batchCount int

	for batch := range batchesCh {
		batchCount++
		totalReceived += len(batch)
	}

	if totalReceived != numStocks {
		t.Errorf("esperado receber %d stocks, recebeu %d", numStocks, totalReceived)
	}

	expectedBatches := 10
	if batchCount != expectedBatches {
		t.Errorf("esperado %d batches, recebeu %d", expectedBatches, batchCount)
	}
}
