package ingest

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/v-venes/backend-challenge/internal/domain"
)

type NewIngestServiceParams struct {
	BatchSize int
	Workers   int
}

type IngestService struct {
	batchSize int
	workers   int
}

func NewIngestService(params NewIngestServiceParams) *IngestService {
	return &IngestService{
		batchSize: params.BatchSize,
		workers:   params.Workers,
	}
}

func (s *IngestService) Run(path string) error {
	stocksCh := make(chan domain.Stock)
	batchesCh := make(chan []domain.Stock)

	reader := NewCSVReader(';')
	batcher := NewBatcher(s.batchSize)

	go batcher.Run(stocksCh, batchesCh)

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		worker := NewBatchWorker()
		wg.Add(1)
		go worker.Run(&wg, batchesCh)
	}

	if err := s.processFiles(path, reader, stocksCh); err != nil {
		close(stocksCh)
		return err
	}

	close(stocksCh)
	wg.Wait()
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

		log.Printf("Reading file %s", filename)

		filepath := fmt.Sprintf("%s/%s", path, filename)
		readCloser, err := zip.OpenReader(filepath)
		if err != nil {
			return err
		}

		for _, file := range readCloser.File {
			log.Printf("Processing file %s inside zip", file.Name)

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
