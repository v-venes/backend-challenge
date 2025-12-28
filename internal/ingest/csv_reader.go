package ingest

import (
	"encoding/csv"
	"io"
	"log"
	"strings"

	"github.com/v-venes/backend-challenge/internal/domain"
)

type CSVReader struct {
	delimiter rune
}

func NewCSVReader(delimiter rune) *CSVReader {
	return &CSVReader{
		delimiter: delimiter,
	}
}

func (r *CSVReader) Read(reader io.Reader, stocksCh chan<- domain.Stock) error {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = r.delimiter

	headers, err := csvReader.Read()
	if err != nil {
		return err
	}

	index := make(map[string]int)
	for i, header := range headers {
		index[strings.ToLower(header)] = i
	}

	line := 2

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			return nil
		} else if err != nil {
			log.Printf("[ERROR] reading csv line=%d err=:%v", line, err)
			line++
			continue
		}

		values := map[string]string{
			"datanegocio":         record[index["datanegocio"]],
			"codigoinstrumento":   record[index["codigoinstrumento"]],
			"preconegocio":        record[index["preconegocio"]],
			"quantidadenegociada": record[index["quantidadenegociada"]],
			"horafechamento":      record[index["horafechamento"]],
		}

		stock, err := domain.StockFromStrArray(values)
		if err != nil {
			log.Printf(
				"[WARN] invalid record line=%d record=%v err=%v",
				line,
				values,
				err,
			)
			line++
			continue
		}

		stocksCh <- *stock
		line++
	}
}
