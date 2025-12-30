package ingest

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/v-venes/backend-challenge/internal/domain"
)

func TestNewCSVReader(t *testing.T) {
	tests := []struct {
		name      string
		delimiter rune
	}{
		{
			name:      "cria reader com delimitador ponto e vírgula",
			delimiter: ';',
		},
		{
			name:      "cria reader com delimitador vírgula",
			delimiter: ',',
		},
		{
			name:      "cria reader com delimitador tab",
			delimiter: '\t',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewCSVReader(tt.delimiter)

			if reader == nil {
				t.Fatal("NewCSVReader retornou nil")
			}

			if reader.delimiter != tt.delimiter {
				t.Errorf("delimiter esperado %q, recebido %q", tt.delimiter, reader.delimiter)
			}
		})
	}
}

func TestCSVReader_Read_Success(t *testing.T) {
	csvData := `DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento
2025-12-01;PETR4;28,50;1000;103000123
2025-12-01;VALE3;65,75;2000;114500123`

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 10)

	go func() {
		err := reader.Read(strings.NewReader(csvData), stocksCh)
		if err != nil {
			t.Errorf("erro inesperado: %v", err)
		}
		close(stocksCh)
	}()

	var stocks []domain.Stock
	for stock := range stocksCh {
		stocks = append(stocks, stock)
	}

	if len(stocks) != 2 {
		t.Errorf("esperado 2 stocks, recebido %d", len(stocks))
	}

	if stocks[0].Ticker != "PETR4" {
		t.Errorf("ticker esperado PETR4, recebido %s", stocks[0].Ticker)
	}
	if stocks[0].Quantity != 1000 {
		t.Errorf("quantity esperado 1000, recebido %d", stocks[0].Quantity)
	}

	if stocks[1].Ticker != "VALE3" {
		t.Errorf("ticker esperado VALE3, recebido %s", stocks[1].Ticker)
	}
	if stocks[1].Quantity != 2000 {
		t.Errorf("quantity esperado 2000, recebido %d", stocks[1].Quantity)
	}
}

func TestCSVReader_Read_EmptyFile(t *testing.T) {
	csvData := `DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento`

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 10)

	go func() {
		err := reader.Read(strings.NewReader(csvData), stocksCh)
		if err != nil {
			t.Errorf("erro inesperado: %v", err)
		}
		close(stocksCh)
	}()

	var stocks []domain.Stock
	for stock := range stocksCh {
		stocks = append(stocks, stock)
	}

	if len(stocks) != 0 {
		t.Errorf("esperado 0 stocks para arquivo vazio, recebido %d", len(stocks))
	}
}

func TestCSVReader_Read_MissingHeaders(t *testing.T) {
	csvData := ``

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 10)

	err := reader.Read(strings.NewReader(csvData), stocksCh)

	if err == nil {
		t.Error("esperado erro para arquivo sem headers")
	}

	if err != io.EOF {
		t.Errorf("esperado io.EOF, recebido %v", err)
	}
}

func TestCSVReader_Read_SkipsInvalidRecords(t *testing.T) {
	csvData := `DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento
2025-12-01;PETR4;28,50;1000;103000123
2025-12-01;INVALID;;;
2025-12-01;VALE3;65,75;2000;114500123
invalid_date;ITUB4;32,10;1500;120000123`

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 10)

	go func() {
		err := reader.Read(strings.NewReader(csvData), stocksCh)
		if err != nil {
			t.Errorf("erro inesperado: %v", err)
		}
		close(stocksCh)
	}()

	var stocks []domain.Stock
	for stock := range stocksCh {
		stocks = append(stocks, stock)
	}

	if len(stocks) != 2 {
		t.Errorf("esperado 2 stocks válidos, recebido %d", len(stocks))
	}

	if stocks[0].Ticker != "PETR4" {
		t.Errorf("primeiro stock deveria ser PETR4, recebido %s", stocks[0].Ticker)
	}

	if stocks[1].Ticker != "VALE3" {
		t.Errorf("segundo stock deveria ser VALE3, recebido %s", stocks[1].Ticker)
	}
}

func TestCSVReader_Read_DifferentDelimiters(t *testing.T) {
	tests := []struct {
		name      string
		delimiter rune
		csvData   string
	}{
		{
			name:      "delimitador tab",
			delimiter: '\t',
			csvData:   "DataNegocio\tCodigoInstrumento\tPrecoNegocio\tQuantidadeNegociada\tHoraFechamento\n2025-12-01\tPETR4\t28,50\t1000\t103000123",
		},
		{
			name:      "delimitador pipe",
			delimiter: '|',
			csvData: `DataNegocio|CodigoInstrumento|PrecoNegocio|QuantidadeNegociada|HoraFechamento
2025-12-01|PETR4|28,50|1000|103000123`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewCSVReader(tt.delimiter)
			stocksCh := make(chan domain.Stock, 10)

			go func() {
				err := reader.Read(strings.NewReader(tt.csvData), stocksCh)
				if err != nil {
					t.Errorf("erro inesperado: %v", err)
				}
				close(stocksCh)
			}()

			var stocks []domain.Stock
			for stock := range stocksCh {
				stocks = append(stocks, stock)
			}

			if len(stocks) != 1 {
				t.Errorf("esperado 1 stock, recebido %d", len(stocks))
			}

			if stocks[0].Ticker != "PETR4" {
				t.Errorf("ticker esperado PETR4, recebido %s", stocks[0].Ticker)
			}
		})
	}
}

func TestCSVReader_Read_CaseInsensitiveHeaders(t *testing.T) {
	tests := []struct {
		name    string
		csvData string
	}{
		{
			name: "headers minúsculos",
			csvData: `datanegocio;codigoinstrumento;preconegocio;quantidadenegociada;horafechamento
2025-12-01;PETR4;28,50;1000;103000123`,
		},
		{
			name: "headers maiúsculos",
			csvData: `DATANEGOCIO;CODIGOINSTRUMENTO;PRECONEGOCIO;QUANTIDADENEGOCIADA;HORAFECHAMENTO
2025-12-01;PETR4;28,50;1000;103000123`,
		},
		{
			name: "headers mixed case",
			csvData: `DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento
2025-12-01;PETR4;28,50;1000;103000123`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewCSVReader(';')
			stocksCh := make(chan domain.Stock, 10)

			go func() {
				err := reader.Read(strings.NewReader(tt.csvData), stocksCh)
				if err != nil {
					t.Errorf("erro inesperado: %v", err)
				}
				close(stocksCh)
			}()

			var stocks []domain.Stock
			for stock := range stocksCh {
				stocks = append(stocks, stock)
			}

			if len(stocks) != 1 {
				t.Errorf("esperado 1 stock, recebido %d", len(stocks))
			}

			if stocks[0].Ticker != "PETR4" {
				t.Errorf("ticker esperado PETR4, recebido %s", stocks[0].Ticker)
			}
		})
	}
}

func TestCSVReader_Read_LargeFile(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento\n")

	expectedCount := 1000
	for i := 0; i < expectedCount; i++ {
		sb.WriteString("2025-12-01;PETR4;28,50;1000;103000123\n")
	}

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 100)

	go func() {
		err := reader.Read(strings.NewReader(sb.String()), stocksCh)
		if err != nil {
			t.Errorf("erro inesperado: %v", err)
		}
		close(stocksCh)
	}()

	count := 0
	for range stocksCh {
		count++
	}

	if count != expectedCount {
		t.Errorf("esperado %d stocks, recebido %d", expectedCount, count)
	}
}

func TestCSVReader_Read_MalformedCSV(t *testing.T) {
	csvData := `DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento
2025-12-01;PETR4;28,50;1000;103000123
2025-12-01;VALE3;65,75
2025-12-01;ITUB4;32,10;1500;120000123`

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 10)

	go func() {
		err := reader.Read(strings.NewReader(csvData), stocksCh)
		if err != nil {
			t.Errorf("erro inesperado: %v", err)
		}
		close(stocksCh)
	}()

	var stocks []domain.Stock
	for stock := range stocksCh {
		stocks = append(stocks, stock)
	}

	if len(stocks) != 2 {
		t.Errorf("esperado 2 stocks válidos, recebido %d", len(stocks))
	}
}

func TestCSVReader_Read_ChannelBuffer(t *testing.T) {
	csvData := `DataNegocio;CodigoInstrumento;PrecoNegocio;QuantidadeNegociada;HoraFechamento
2025-12-01;PETR4;28,50;1000;103000123
2025-12-01;VALE3;65,75;2000;114500123
2025-12-01;ITUB4;32,10;1500;120000123`

	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock)

	go func() {
		err := reader.Read(strings.NewReader(csvData), stocksCh)
		if err != nil {
			t.Errorf("erro inesperado: %v", err)
		}
		close(stocksCh)
	}()

	var stocks []domain.Stock
	for stock := range stocksCh {
		time.Sleep(10 * time.Millisecond)
		stocks = append(stocks, stock)
	}

	if len(stocks) != 3 {
		t.Errorf("esperado 3 stocks, recebido %d", len(stocks))
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error simulado")
}

func TestCSVReader_Read_ReaderError(t *testing.T) {
	reader := NewCSVReader(';')
	stocksCh := make(chan domain.Stock, 10)

	err := reader.Read(&errorReader{}, stocksCh)

	if err == nil {
		t.Error("esperado erro de leitura")
	}
}
