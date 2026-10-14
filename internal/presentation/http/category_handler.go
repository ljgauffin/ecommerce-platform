package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
)

type CategoryHandler struct {
	categoryRepo domain.CategoryRepository
}

func NewCategoryHandler(categoryRepo domain.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

type CategoryInput struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryRepo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if categories == nil {
		categories = []domain.Category{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": categories,
	})
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	trimmed := strings.TrimSpace(input.Name)
	if trimmed == "" {
		http.Error(w, "category name cannot be empty", http.StatusBadRequest)
		return
	}

	cat, err := domain.NewCategory(trimmed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.categoryRepo.Create(r.Context(), cat); err != nil {
		if errors.Is(err, domain.ErrCategoryDuplicate) {
			http.Error(w, "category already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cat)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.categoryRepo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrCategoryHasProducts) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
