package usecase

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/sanitizer"
)

type RowError struct {
	Row   int    `json:"row"`
	SKU   string `json:"sku,omitempty"`
	Field string `json:"field"`
	Error string `json:"error"`
}

type ImportResult struct {
	TotalRows int        `json:"total_rows"`
	Created   int        `json:"created"`
	Updated   int        `json:"updated"`
	Imported  int        `json:"imported"`
	Failed    int        `json:"failed"`
	Errors    []RowError `json:"errors"`
}

type CsvImportUseCase struct {
	repo domain.ProductRepository
}

func NewCsvImportUseCase(repo domain.ProductRepository) *CsvImportUseCase {
	return &CsvImportUseCase{repo: repo}
}

func (uc *CsvImportUseCase) Execute(ctx context.Context, reader io.Reader) (*ImportResult, error) {
	normalizedReader := newCRStripper(reader)

	bufReader := bufio.NewReader(normalizedReader)
	preview, err := bufReader.Peek(1024)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("read preview: %w", err)
	}

	delimiter := rune(',')
	if bytes.Contains(preview, []byte("\t")) && !bytes.Contains(preview, []byte(",")) {
		delimiter = '\t'
	}

	csvReader := csv.NewReader(bufReader)
	csvReader.Comma = delimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read headers: %w", err)
	}

	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	requiredCols := []string{"name", "sku", "category", "price", "stock", "weight_kg"}
	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			return nil, fmt.Errorf("missing required column in CSV header: %s", col)
		}
	}

	result := &ImportResult{
		Errors: make([]RowError, 0),
	}

	type rowCandidate struct {
		rowNumber int
		product   domain.Product
	}

	skuOccurrences := make(map[string][]rowCandidate)
	var candidates []rowCandidate
	rowNumber := 1

	for {
		record, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		rowNumber++
		result.TotalRows++

		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				Field: "row",
				Error: fmt.Sprintf("malformed CSV line: %v", err),
			})
			continue
		}

		getCol := func(colName string) string {
			if idx, ok := colMap[colName]; ok && idx < len(record) {
				return record[idx]
			}
			return ""
		}

		rawSKU := getCol("sku")
		sku, err := sanitizer.CleanString(rawSKU)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				Field: "sku",
				Error: "SKU cannot be empty or whitespace only",
			})
			continue
		}

		rawName := getCol("name")
		name, err := sanitizer.CleanString(rawName)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				SKU:   sku,
				Field: "name",
				Error: "product name cannot be empty or whitespace only",
			})
			continue
		}

		description := sanitizer.SanitizeText(getCol("description"))

		rawCategory := getCol("category")
		category, err := sanitizer.CleanString(rawCategory)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				SKU:   sku,
				Field: "category",
				Error: "category cannot be empty or whitespace only",
			})
			continue
		}

		price, err := sanitizer.ParsePrice(getCol("price"))
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				SKU:   sku,
				Field: "price",
				Error: fmt.Sprintf("invalid price '%s': %v", getCol("price"), err),
			})
			continue
		}

		stock, err := sanitizer.ParseStock(getCol("stock"))
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				SKU:   sku,
				Field: "stock",
				Error: fmt.Sprintf("invalid stock '%s': %v", getCol("stock"), err),
			})
			continue
		}

		weight, err := sanitizer.ParseWeight(getCol("weight_kg"))
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				SKU:   sku,
				Field: "weight_kg",
				Error: fmt.Sprintf("invalid weight '%s': %v", getCol("weight_kg"), err),
			})
			continue
		}

		prod, err := domain.NewProduct(sku, name, description, category, price, stock, weight)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   rowNumber,
				SKU:   sku,
				Field: "domain",
				Error: err.Error(),
			})
			continue
		}

		cand := rowCandidate{
			rowNumber: rowNumber,
			product:   *prod,
		}
		skuOccurrences[sku] = append(skuOccurrences[sku], cand)
		candidates = append(candidates, cand)
	}

	var validProducts []domain.Product
	for _, cand := range candidates {
		occurrences := skuOccurrences[cand.product.SKU]
		if len(occurrences) > 1 {
			result.Failed++
			result.Errors = append(result.Errors, RowError{
				Row:   cand.rowNumber,
				SKU:   cand.product.SKU,
				Field: "sku",
				Error: fmt.Sprintf("duplicate SKU in CSV: SKU '%s' appears %d times, all occurrences rejected", cand.product.SKU, len(occurrences)),
			})
		} else {
			validProducts = append(validProducts, cand.product)
		}
	}

	sort.Slice(result.Errors, func(i, j int) bool {
		return result.Errors[i].Row < result.Errors[j].Row
	})

	if len(validProducts) > 0 {
		created, updated, err := uc.repo.BulkUpsert(ctx, validProducts)
		if err != nil {
			return nil, fmt.Errorf("bulk upsert valid products: %w", err)
		}
		result.Created = created
		result.Updated = updated
		result.Imported = created + updated
	}

	return result, nil
}

type crStripper struct {
	r io.Reader
}

func newCRStripper(r io.Reader) io.Reader {
	return &crStripper{r: r}
}

func (s *crStripper) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	w := 0
	for i := 0; i < n; i++ {
		if p[i] != '\r' {
			p[w] = p[i]
			w++
		}
	}
	return w, err
}
