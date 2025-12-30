package ingest

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/v-venes/backend-challenge/internal/domain"
)

type MockStockRepository struct {
	UpsertBatchFunc func(stocks []domain.Stock) error
	calls           [][]domain.Stock
	mu              sync.Mutex
}

func (m *MockStockRepository) UpsertBatch(stocks []domain.Stock) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, stocks)

	if m.UpsertBatchFunc != nil {
		return m.UpsertBatchFunc(stocks)
	}
	return nil
}

func (m *MockStockRepository) GetAggregatedByTicker(ticker string, startDate time.Time, endDate time.Time) (*domain.AggregatedStockResult, error) {
	return nil, nil
}

func (m *MockStockRepository) GetCalls() [][]domain.Stock {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func (m *MockStockRepository) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func TestNewBatchWorker(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	if worker == nil {
		t.Fatal("NewBatchWorker retornou nil")
	}

	if worker.repo == nil {
		t.Error("repository não foi inicializado")
	}
}

func TestBatchWorker_Run_SingleBatch(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 1)

	batch := []domain.Stock{
		{Ticker: "PETR4", Quantity: 100},
		{Ticker: "VALE3", Quantity: 200},
	}

	batchesCh <- batch
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	worker.Run(&wg, batchesCh)

	if mockRepo.CallCount() != 1 {
		t.Errorf("esperado 1 chamada ao UpsertBatch, recebeu %d", mockRepo.CallCount())
	}

	calls := mockRepo.GetCalls()
	if len(calls[0]) != 2 {
		t.Errorf("esperado batch com 2 stocks, recebeu %d", len(calls[0]))
	}

	if calls[0][0].Ticker != "PETR4" {
		t.Errorf("esperado ticker PETR4, recebeu %s", calls[0][0].Ticker)
	}
}

func TestBatchWorker_Run_MultipleBatches(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 5)

	batches := [][]domain.Stock{
		{{Ticker: "PETR4", Quantity: 100}, {Ticker: "VALE3", Quantity: 200}},
		{{Ticker: "ITUB4", Quantity: 300}},
		{{Ticker: "BBDC4", Quantity: 400}, {Ticker: "ABEV3", Quantity: 500}, {Ticker: "WEGE3", Quantity: 600}},
	}

	for _, batch := range batches {
		batchesCh <- batch
	}
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	worker.Run(&wg, batchesCh)

	if mockRepo.CallCount() != 3 {
		t.Errorf("esperado 3 chamadas ao UpsertBatch, recebeu %d", mockRepo.CallCount())
	}

	calls := mockRepo.GetCalls()

	if len(calls[0]) != 2 {
		t.Errorf("batch 1: esperado 2 stocks, recebeu %d", len(calls[0]))
	}

	if len(calls[1]) != 1 {
		t.Errorf("batch 2: esperado 1 stock, recebeu %d", len(calls[1]))
	}

	if len(calls[2]) != 3 {
		t.Errorf("batch 3: esperado 3 stocks, recebeu %d", len(calls[2]))
	}
}

func TestBatchWorker_Run_EmptyChannel(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock)
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	worker.Run(&wg, batchesCh)

	if mockRepo.CallCount() != 0 {
		t.Errorf("esperado 0 chamadas para canal vazio, recebeu %d", mockRepo.CallCount())
	}
}

func TestBatchWorker_Run_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	mockRepo := &MockStockRepository{
		UpsertBatchFunc: func(stocks []domain.Stock) error {
			return expectedErr
		},
	}

	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 1)
	batchesCh <- []domain.Stock{{Ticker: "PETR4", Quantity: 100}}
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	worker.Run(&wg, batchesCh)

	if mockRepo.CallCount() != 1 {
		t.Errorf("esperado 1 chamada ao UpsertBatch, recebeu %d", mockRepo.CallCount())
	}
}

func TestBatchWorker_Run_WaitGroupDecrement(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 1)
	batchesCh <- []domain.Stock{{Ticker: "PETR4", Quantity: 100}}
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	done := make(chan bool)

	go func() {
		worker.Run(&wg, batchesCh)
	}()

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:

	case <-time.After(1 * time.Second):
		t.Error("timeout: WaitGroup não foi decrementado")
	}
}

func TestBatchWorker_Run_ConcurrentWorkers(t *testing.T) {
	mockRepo := &MockStockRepository{}

	batchesCh := make(chan []domain.Stock, 10)

	for i := 0; i < 10; i++ {
		batch := []domain.Stock{
			{Ticker: "TEST", Quantity: int64(i)},
		}
		batchesCh <- batch
	}
	close(batchesCh)

	var wg sync.WaitGroup
	numWorkers := 3

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		worker := NewBatchWorker(mockRepo)
		go worker.Run(&wg, batchesCh)
	}

	wg.Wait()

	if mockRepo.CallCount() != 10 {
		t.Errorf("esperado 10 chamadas ao UpsertBatch, recebeu %d", mockRepo.CallCount())
	}
}

func TestBatchWorker_Run_LargeBatches(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 1)

	largeBatch := make([]domain.Stock, 1000)
	for i := 0; i < 1000; i++ {
		largeBatch[i] = domain.Stock{
			Ticker:   "PETR4",
			Quantity: int64(i),
		}
	}

	batchesCh <- largeBatch
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	worker.Run(&wg, batchesCh)

	if mockRepo.CallCount() != 1 {
		t.Errorf("esperado 1 chamada, recebeu %d", mockRepo.CallCount())
	}

	calls := mockRepo.GetCalls()
	if len(calls[0]) != 1000 {
		t.Errorf("esperado batch com 1000 stocks, recebeu %d", len(calls[0]))
	}
}

func TestBatchWorker_Run_ProcessingOrder(t *testing.T) {
	mockRepo := &MockStockRepository{}
	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 3)

	batch1 := []domain.Stock{{Ticker: "BATCH1", Quantity: 1}}
	batch2 := []domain.Stock{{Ticker: "BATCH2", Quantity: 2}}
	batch3 := []domain.Stock{{Ticker: "BATCH3", Quantity: 3}}

	batchesCh <- batch1
	batchesCh <- batch2
	batchesCh <- batch3
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	worker.Run(&wg, batchesCh)

	calls := mockRepo.GetCalls()

	if calls[0][0].Ticker != "BATCH1" {
		t.Errorf("primeiro batch esperado BATCH1, recebeu %s", calls[0][0].Ticker)
	}

	if calls[1][0].Ticker != "BATCH2" {
		t.Errorf("segundo batch esperado BATCH2, recebeu %s", calls[1][0].Ticker)
	}

	if calls[2][0].Ticker != "BATCH3" {
		t.Errorf("terceiro batch esperado BATCH3, recebeu %s", calls[2][0].Ticker)
	}
}

func TestBatchWorker_Run_SlowRepository(t *testing.T) {
	mockRepo := &MockStockRepository{
		UpsertBatchFunc: func(stocks []domain.Stock) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}

	worker := NewBatchWorker(mockRepo)

	batchesCh := make(chan []domain.Stock, 3)

	for i := 0; i < 3; i++ {
		batchesCh <- []domain.Stock{{Ticker: "SLOW", Quantity: int64(i)}}
	}
	close(batchesCh)

	var wg sync.WaitGroup
	wg.Add(1)

	start := time.Now()
	worker.Run(&wg, batchesCh)
	elapsed := time.Since(start)

	if elapsed < 150*time.Millisecond {
		t.Errorf("processamento muito rápido: %v", elapsed)
	}

	if mockRepo.CallCount() != 3 {
		t.Errorf("esperado 3 chamadas, recebeu %d", mockRepo.CallCount())
	}
}
