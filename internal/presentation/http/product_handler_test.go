package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	apphttp "github.com/lucasgauffin/ecommerce-platform/internal/presentation/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestRouter(t *testing.T) (*chi.Mux, domain.ProductRepository) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Category{}, &domain.Product{}, &domain.Order{}, &domain.OrderItem{}, &domain.IdempotencyRecord{})
	require.NoError(t, err)

	repo := persistence.NewGormProductRepo(db)
	catRepo := persistence.NewGormCategoryRepo(db)
	handler := apphttp.NewProductHandler(repo, catRepo)
	catHandler := apphttp.NewCategoryHandler(catRepo)
	swaggerHandler := apphttp.NewSwaggerHandler()

	r := chi.NewRouter()
	r.Get("/swagger", swaggerHandler.UI)
	r.Get("/swagger/doc.json", swaggerHandler.Doc)

	r.Route("/api/v1/categories", func(r chi.Router) {
		r.Get("/", catHandler.List)
		r.Post("/", catHandler.Create)
		r.Delete("/{id}", catHandler.Delete)
	})

	r.Route("/api/v1/products", func(r chi.Router) {
		r.Get("/", handler.List)
		r.Get("/categories", handler.GetCategories)
		r.Post("/", handler.Create)
		r.Get("/{id}", handler.GetByID)
		r.Put("/{id}", handler.Update)
		r.Delete("/{id}", handler.Delete)
	})

	return r, repo
}

func TestProductHandler_CRUD(t *testing.T) {
	r, _ := setupTestRouter(t)

	catBody := []byte(`{"name":"Electronics"}`)
	req, _ := http.NewRequest("POST", "/api/v1/categories", bytes.NewBuffer(catBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	invalidCatBody := []byte(`{
		"sku": "BAD-001",
		"name": "Bad Product",
		"category": "NonExistentCategory",
		"price": 10.0,
		"stock": 5,
		"weight_kg": 0.1
	}`)
	req, _ = http.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(invalidCatBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	createBody := []byte(`{
		"sku": "SWAG-001",
		"name": "Smart Watch",
		"description": "Fitness tracking",
		"category": "Electronics",
		"price": 149.99,
		"stock": 50,
		"weight_kg": 0.15
	}`)

	req, _ = http.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var created domain.Product
	err := json.Unmarshal(w.Body.Bytes(), &created)
	require.NoError(t, err)
	assert.Equal(t, "SWAG-001", created.SKU)
	assert.NotEmpty(t, created.ID)

	req, _ = http.NewRequest("GET", "/api/v1/products?query=watch&category=Electronics&in_stock=true&sort=price-asc", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var listResp struct {
		Data  []domain.Product `json:"data"`
		Total int64            `json:"total"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)
	assert.Equal(t, int64(1), listResp.Total)
	assert.Len(t, listResp.Data, 1)

	req, _ = http.NewRequest("GET", "/api/v1/products/categories", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var catResp struct {
		Data []string `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &catResp)
	require.NoError(t, err)
	assert.Contains(t, catResp.Data, "Electronics")

	req, _ = http.NewRequest("GET", "/api/v1/products/"+created.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req, _ = http.NewRequest("DELETE", "/api/v1/products/"+created.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	req, _ = http.NewRequest("GET", "/api/v1/products/"+created.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSwaggerHandler_Endpoints(t *testing.T) {
	r, _ := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/swagger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "SwaggerUIBundle")

	req, _ = http.NewRequest("GET", "/swagger/doc.json", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Enterprise E-Commerce Platform API")
}
