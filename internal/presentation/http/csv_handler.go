package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
)

type CsvHandler struct {
	importUC *usecase.CsvImportUseCase
}

func NewCsvHandler(importUC *usecase.CsvImportUseCase) *CsvHandler {
	return &CsvHandler{importUC: importUC}
}

func (h *CsvHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	var reader io.Reader

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Error(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing 'file' field in form-data", http.StatusBadRequest)
			return
		}
		defer file.Close()
		reader = file
	} else {
		reader = r.Body
	}

	result, err := h.importUC.Execute(r.Context(), reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
