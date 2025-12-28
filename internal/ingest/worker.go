package ingest

import (
	"sync"

	"github.com/v-venes/backend-challenge/internal/domain"
	"github.com/v-venes/backend-challenge/internal/repository"
)

type BatchWorker struct {
	repo repository.StockRepository
}

func NewBatchWorker(repo repository.StockRepository) *BatchWorker {
	return &BatchWorker{
		repo: repo,
	}
}

func (b *BatchWorker) Run(wg *sync.WaitGroup, batchesCh <-chan []domain.Stock) {
	defer wg.Done()

	for batch := range batchesCh {
		b.repo.UpsertBatch(batch)
	}
}
