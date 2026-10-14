package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
)

type ProductHandler struct {
	repo         domain.ProductRepository
	categoryRepo domain.CategoryRepository
}

func NewProductHandler(repo domain.ProductRepository, categoryRepo domain.CategoryRepository) *ProductHandler {
	return &ProductHandler{repo: repo, categoryRepo: categoryRepo}
}

type ProductInput struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	WeightKg    float64 `json:"weight_kg"`
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var cat *domain.Category
	if h.categoryRepo != nil {
		if strings.TrimSpace(input.CategoryID) != "" {
			c, err := h.categoryRepo.GetByID(r.Context(), input.CategoryID)
			if err != nil {
				if errors.Is(err, domain.ErrCategoryNotFound) {
					http.Error(w, "category_id '"+input.CategoryID+"' does not exist; please select a valid category", http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cat = c
		} else if strings.TrimSpace(input.Category) != "" {
			c, err := h.categoryRepo.GetByName(r.Context(), input.Category)
			if err != nil {
				if errors.Is(err, domain.ErrCategoryNotFound) {
					http.Error(w, "category '"+input.Category+"' does not exist; please select an existing category or create it first", http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cat = c
		}
	}

	var categoryID string
	if cat != nil {
		categoryID = cat.ID
	} else if strings.TrimSpace(input.CategoryID) != "" {
		categoryID = input.CategoryID
	}

	prod, err := domain.NewProduct(
		input.SKU,
		input.Name,
		input.Description,
		categoryID,
		input.Price,
		input.Stock,
		input.WeightKg,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	prod.Category = cat

	if err := h.repo.Create(r.Context(), prod); err != nil {
		if errors.Is(err, domain.ErrProductDuplicateSKU) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prod)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("query")
	if search == "" {
		search = q.Get("search")
	}
	sortVal := q.Get("sort")
	if sortVal == "" {
		sortVal = q.Get("sort_by")
	}

	filter := domain.ProductFilter{
		Query:       search,
		Category:    q.Get("category"),
		CategoryID:  q.Get("category_id"),
		InStockOnly: q.Get("in_stock") == "true",
		SortBy:      sortVal,
		Limit:       20,
		Offset:      0,
	}

	if minPriceStr := q.Get("min_price"); minPriceStr != "" {
		if val, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = &val
		}
	}
	if maxPriceStr := q.Get("max_price"); maxPriceStr != "" {
		if val, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = &val
		}
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			if val > 1000 {
				val = 1000
			}
			filter.Limit = val
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	products, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":   products,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	if h.categoryRepo != nil {
		cats, err := h.categoryRepo.List(r.Context())
		if err == nil && len(cats) > 0 {
			names := make([]string, len(cats))
			for i, c := range cats {
				names[i] = c.Name
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": names,
			})
			return
		}
	}

	categories, err := h.repo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if categories == nil {
		categories = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": categories,
	})
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	prod, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prod)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	prod, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var input ProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if h.categoryRepo != nil {
		if strings.TrimSpace(input.CategoryID) != "" {
			cat, err := h.categoryRepo.GetByID(r.Context(), strings.TrimSpace(input.CategoryID))
			if err != nil {
				if errors.Is(err, domain.ErrCategoryNotFound) {
					http.Error(w, "category_id '"+input.CategoryID+"' does not exist; please select a valid category", http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			prod.CategoryID = cat.ID
			prod.Category = cat
		} else if strings.TrimSpace(input.Category) != "" {
			cat, err := h.categoryRepo.GetByName(r.Context(), strings.TrimSpace(input.Category))
			if err != nil {
				if errors.Is(err, domain.ErrCategoryNotFound) {
					http.Error(w, "category '"+input.Category+"' does not exist; please select an existing category or create it first", http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			prod.CategoryID = cat.ID
			prod.Category = cat
		}
	}

	prod.SKU = strings.TrimSpace(input.SKU)
	prod.Name = strings.TrimSpace(input.Name)
	prod.Description = strings.TrimSpace(input.Description)
	prod.Price = input.Price
	prod.Stock = input.Stock
	prod.WeightKg = input.WeightKg

	if err := prod.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Update(r.Context(), prod); err != nil {
		if errors.Is(err, domain.ErrProductDuplicateSKU) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prod)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
