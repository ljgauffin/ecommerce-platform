package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	apphttp "github.com/lucasgauffin/ecommerce-platform/internal/presentation/http"
	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupOrderHandlerTest(t *testing.T) (http.Handler, domain.ProductRepository) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&domain.Category{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.IdempotencyRecord{},
	)
	require.NoError(t, err)

	productRepo := persistence.NewGormProductRepo(db)
	orderRepo := persistence.NewGormOrderRepo(db)
	idempRepo := persistence.NewGormIdempotencyRepo(db)
	checkoutUC := usecase.NewCheckoutUseCase(productRepo, orderRepo, idempRepo)
	orderHandler := apphttp.NewOrderHandler(checkoutUC)

	r := chi.NewRouter()
	r.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/checkout", orderHandler.Checkout)
		r.Get("/{id}", orderHandler.GetByID)
	})

	return r, productRepo
}

func TestOrderHandler_CheckoutAndGet(t *testing.T) {
	router, productRepo := setupOrderHandlerTest(t)
	ctx := context.Background()

	prod, err := domain.NewProduct("SKU-ORD-1", "Test Product", "Description", "Category", 25.0, 10, 0.5)
	require.NoError(t, err)
	require.NoError(t, productRepo.Create(ctx, prod))

	body := map[string]interface{}{
		"customer_email": "test@buyer.com",
		"items": []map[string]interface{}{
			{"product_id": prod.ID, "quantity": 2},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/orders/checkout", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "http-idemp-001")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var createdOrder domain.Order
	err = json.Unmarshal(rec.Body.Bytes(), &createdOrder)
	require.NoError(t, err)
	assert.NotEmpty(t, createdOrder.ID)
	assert.Equal(t, 50.0, createdOrder.TotalAmount)

	getReq := httptest.NewRequest("GET", "/api/v1/orders/"+createdOrder.ID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	assert.Equal(t, http.StatusOK, getRec.Code)

	replayReq := httptest.NewRequest("POST", "/api/v1/orders/checkout", bytes.NewReader(bodyBytes))
	replayReq.Header.Set("Content-Type", "application/json")
	replayReq.Header.Set("Idempotency-Key", "http-idemp-001")
	replayRec := httptest.NewRecorder()

	router.ServeHTTP(replayRec, replayReq)
	assert.Equal(t, http.StatusCreated, replayRec.Code)

	var replayedOrder domain.Order
	err = json.Unmarshal(replayRec.Body.Bytes(), &replayedOrder)
	require.NoError(t, err)
	assert.Equal(t, createdOrder.ID, replayedOrder.ID)

	largeBody := map[string]interface{}{
		"customer_email": "test@buyer.com",
		"items": []map[string]interface{}{
			{"product_id": prod.ID, "quantity": 100},
		},
	}
	largeBytes, _ := json.Marshal(largeBody)

	failReq := httptest.NewRequest("POST", "/api/v1/orders/checkout", bytes.NewReader(largeBytes))
	failReq.Header.Set("Content-Type", "application/json")
	failRec := httptest.NewRecorder()

	router.ServeHTTP(failRec, failReq)
	assert.Equal(t, http.StatusConflict, failRec.Code)
}
