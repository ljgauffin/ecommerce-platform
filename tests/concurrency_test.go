package tests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getTestPostgresDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=ecommerce password=ecommerce_secret dbname=ecommerce_db port=5433 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("PostgreSQL not available at %s, skipping live concurrency test: %v", dsn, err)
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil || sqlDB.Ping() != nil {
		t.Skipf("PostgreSQL ping failed, skipping live concurrency test")
		return nil
	}

	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(10)

	err = db.AutoMigrate(
		&domain.Category{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.IdempotencyRecord{},
	)
	require.NoError(t, err)

	return db
}

func TestConcurrentCheckout_StrictRowLevelLocking(t *testing.T) {
	db := getTestPostgresDB(t)
	productRepo := persistence.NewGormProductRepo(db)
	orderRepo := persistence.NewGormOrderRepo(db)
	idempRepo := persistence.NewGormIdempotencyRepo(db)
	checkoutUC := usecase.NewCheckoutUseCase(productRepo, orderRepo, idempRepo)

	ctx := context.Background()

	uniqueSKU := fmt.Sprintf("CONC-%s", uuid.NewString()[:8])
	initialStock := 5
	prod, err := domain.NewProduct(uniqueSKU, "Flash Sale Item", "High contention item", "Specials", 15.0, initialStock, 0.5)
	require.NoError(t, err)
	require.NoError(t, productRepo.Create(ctx, prod))

	totalBuyers := 20
	var successCount int64
	var failureCount int64
	var successfulKey string
	var successfulKeyMu sync.Mutex

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for i := 0; i < totalBuyers; i++ {
		wg.Add(1)
		buyerEmail := fmt.Sprintf("buyer-%d@test.com", i)
		idempKey := fmt.Sprintf("idemp-conc-%s-%d", uniqueSKU, i)

		go func() {
			defer wg.Done()
			<-startSignal 

			order, err := checkoutUC.Execute(context.Background(), usecase.CheckoutInput{
				CustomerEmail:  buyerEmail,
				Items:          []usecase.CheckoutItemInput{{ProductID: prod.ID, Quantity: 1}},
				IdempotencyKey: idempKey,
			})

			if err == nil && order != nil {
				atomic.AddInt64(&successCount, 1)
				successfulKeyMu.Lock()
				if successfulKey == "" {
					successfulKey = idempKey
				}
				successfulKeyMu.Unlock()
			} else {
				atomic.AddInt64(&failureCount, 1)
			}
		}()
	}

	close(startSignal)
	wg.Wait()

	assert.Equal(t, int64(initialStock), atomic.LoadInt64(&successCount), "Only exactly initialStock buyers should succeed")
	assert.Equal(t, int64(totalBuyers-initialStock), atomic.LoadInt64(&failureCount), "Remaining buyers must fail with insufficient stock")

	finalProduct, err := productRepo.GetByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, finalProduct.Stock, "Final stock in PostgreSQL must be exactly 0")

	require.NotEmpty(t, successfulKey, "At least one successful buyer key should be recorded")
	replayedOrder, err := checkoutUC.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail:  "buyer-replay@test.com",
		Items:          []usecase.CheckoutItemInput{{ProductID: prod.ID, Quantity: 1}},
		IdempotencyKey: successfulKey,
	})
	require.NoError(t, err)
	require.NotNil(t, replayedOrder)

	finalProductAfterReplay, err := productRepo.GetByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, finalProductAfterReplay.Stock, "Stock must remain 0 after idempotent replay")
}
