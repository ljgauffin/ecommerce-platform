package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	return db
}

func TestGormProductRepo_CRUD_And_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := persistence.NewGormProductRepo(db)
	ctx := context.Background()

	p, err := domain.NewProduct("SKU-100", "Wireless Headphones", "Noise cancelling", "Electronics", 199.99, 50, 0.25)
	require.NoError(t, err)

	err = repo.Create(ctx, p)
	require.NoError(t, err)

	err = repo.Create(ctx, p)
	assert.ErrorIs(t, err, domain.ErrProductDuplicateSKU)

	fetched, err := repo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "Wireless Headphones", fetched.Name)

	p.Name = "Wireless Headphones Pro"
	err = repo.Update(ctx, p)
	require.NoError(t, err)

	updated, err := repo.GetBySKU(ctx, p.SKU)
	require.NoError(t, err)
	assert.Equal(t, "Wireless Headphones Pro", updated.Name)

	products, total, err := repo.List(ctx, domain.ProductFilter{Query: "headphones"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, products, 1)

	err = repo.Delete(ctx, p.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, p.ID)
	assert.ErrorIs(t, err, domain.ErrProductNotFound)
}

func TestGormOrderRepo_CreateWithLock(t *testing.T) {
	db := setupTestDB(t)
	productRepo := persistence.NewGormProductRepo(db)
	orderRepo := persistence.NewGormOrderRepo(db)
	ctx := context.Background()

	p, err := domain.NewProduct("SKU-200", "Mechanical Keyboard", "RGB switches", "Electronics", 80.0, 5, 0.8)
	require.NoError(t, err)
	require.NoError(t, productRepo.Create(ctx, p))

	order, err := domain.NewOrder("buyer@test.com", []domain.OrderItem{
		{
			ProductID:   p.ID,
			SKU:         p.SKU,
			ProductName: p.Name,
			UnitPrice:   p.Price,
			Quantity:    2,
		},
	})
	require.NoError(t, err)

	err = orderRepo.CreateWithLock(ctx, order)
	require.NoError(t, err)

	updatedProduct, err := productRepo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, updatedProduct.Stock)

	tooLargeOrder, err := domain.NewOrder("buyer@test.com", []domain.OrderItem{
		{
			ProductID:   p.ID,
			SKU:         p.SKU,
			ProductName: p.Name,
			UnitPrice:   p.Price,
			Quantity:    10,
		},
	})
	require.NoError(t, err)

	err = orderRepo.CreateWithLock(ctx, tooLargeOrder)
	assert.ErrorIs(t, domain.ErrInsufficientStock, err)

	productAfterFail, err := productRepo.GetByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, productAfterFail.Stock)
}

func TestGormIdempotencyRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := persistence.NewGormIdempotencyRepo(db)
	ctx := context.Background()

	key := "test-idem-key-1"
	record := &domain.IdempotencyRecord{
		Key:          key,
		Status:       domain.IdempotencyStatusCompleted,
		ResponseCode: 201,
		ResponseBody: []byte(`{"order_id":"123"}`),
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
	}

	err := repo.Save(ctx, record)
	require.NoError(t, err)

	fetched, err := repo.Get(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, key, fetched.Key)
	assert.Equal(t, domain.IdempotencyStatusCompleted, fetched.Status)
	assert.Equal(t, 201, fetched.ResponseCode)
}

func TestGormCategoryRepo(t *testing.T) {
	db := setupTestDB(t)
	catRepo := persistence.NewGormCategoryRepo(db)
	prodRepo := persistence.NewGormProductRepo(db)
	ctx := context.Background()

	cat, err := domain.NewCategory("Apparel")
	require.NoError(t, err)
	err = catRepo.Create(ctx, cat)
	require.NoError(t, err)

	dupCat, _ := domain.NewCategory("Apparel")
	err = catRepo.Create(ctx, dupCat)
	assert.ErrorIs(t, err, domain.ErrCategoryDuplicate)

	found, err := catRepo.GetByName(ctx, "apparel")
	require.NoError(t, err)
	assert.Equal(t, cat.ID, found.ID)

	autoCat, err := catRepo.FindOrCreate(ctx, "Footwear")
	require.NoError(t, err)
	assert.Equal(t, "Footwear", autoCat.Name)

	list, err := catRepo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 2)

	prod, err := domain.NewProduct("SKU-CAT-1", "Sneakers", "Desc", autoCat.Name, 50.0, 10, 0.5)
	require.NoError(t, err)
	prod.CategoryID = autoCat.ID
	require.NoError(t, prodRepo.Create(ctx, prod))

	err = catRepo.Delete(ctx, autoCat.ID)
	assert.ErrorIs(t, err, domain.ErrCategoryHasProducts)

	err = catRepo.Delete(ctx, cat.ID)
	require.NoError(t, err)
}
