package ingest

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/v-venes/backend-challenge/internal/domain"
	"github.com/v-venes/backend-challenge/internal/repository"
)

type NewIngestServiceParams struct {
	BatchSize int
	Workers   int
	Repo      repository.StockRepository
}

type IngestService struct {
	batchSize int
	workers   int
	repo      repository.StockRepository
}

func NewIngestService(params NewIngestServiceParams) *IngestService {
	return &IngestService{
		batchSize: params.BatchSize,
		workers:   params.Workers,
		repo:      params.Repo,
	}
}

func (s *IngestService) Run(path string) error {
	log.Println("[INFO] starting ingestion")
	start := time.Now()
	stocksCh := make(chan domain.Stock)
	batchesCh := make(chan []domain.Stock)

	reader := NewCSVReader(';')
	batcher := NewBatcher(s.batchSize)

	go batcher.Run(stocksCh, batchesCh)

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		worker := NewBatchWorker(s.repo)
		wg.Add(1)
		go worker.Run(&wg, batchesCh)
	}

	if err := s.processFiles(path, reader, stocksCh); err != nil {
		close(stocksCh)
		return err
	}

	close(stocksCh)
	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("[INFO] ingestion finished in %s", elapsed.String())
	return nil
}

func (s *IngestService) processFiles(path string, csvReader *CSVReader, stocksCh chan<- domain.Stock) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if !strings.HasSuffix(filename, ".zip") || file.IsDir() {
			continue
		}

		log.Printf("[INFO] reading file %s", filename)

		filepath := fmt.Sprintf("%s/%s", path, filename)
		readCloser, err := zip.OpenReader(filepath)
		if err != nil {
			return err
		}

		for _, file := range readCloser.File {
			log.Printf("[INFO] processing file %s inside zip", file.Name)

			fileReadCloser, err := file.Open()
			if err != nil {
				return err
			}

			if err = csvReader.Read(fileReadCloser, stocksCh); err != nil {
				return err
			}

			fileReadCloser.Close()
		}

		readCloser.Close()
	}

	return nil
}
