package ingest

import "github.com/v-venes/backend-challenge/internal/domain"

type Batcher struct {
	size  int
	batch []domain.Stock
}

func NewBatcher(batchSize int) *Batcher {
	return &Batcher{
		size:  batchSize,
		batch: make([]domain.Stock, 0, batchSize),
	}
}

func (b *Batcher) Run(stocksCh <-chan domain.Stock, batchesCh chan<- []domain.Stock) {

	for stock := range stocksCh {
		b.batch = append(b.batch, stock)

		if len(b.batch) == b.size {
			batchesCh <- b.batch
			b.batch = make([]domain.Stock, 0, b.size)
		}
	}

	if len(b.batch) > 0 {
		batchesCh <- b.batch
	}

	close(batchesCh)
}
