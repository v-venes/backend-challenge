package ingest

import (
	"log"
	"sync"

	"github.com/v-venes/backend-challenge/internal/domain"
)

type BatchWorker struct{}

func NewBatchWorker() *BatchWorker {
	//TODO: Passar o repo
	return &BatchWorker{}
}

func (b *BatchWorker) Run(wg *sync.WaitGroup, batchesCh <-chan []domain.Stock) {
	defer wg.Done()

	for batch := range batchesCh {
		log.Println("Saving Batch", len(batch))

	}

}
