package usecase_test

import (
	"context"
	"testing"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/infrastructure/persistence"
	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCheckoutTestDB(t *testing.T) (*gorm.DB, domain.ProductRepository, domain.OrderRepository, domain.IdempotencyRepository) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.IdempotencyRecord{},
	)
	require.NoError(t, err)

	return db, persistence.NewGormProductRepo(db), persistence.NewGormOrderRepo(db), persistence.NewGormIdempotencyRepo(db)
}

func TestCheckoutUseCase_Success(t *testing.T) {
	_, productRepo, orderRepo, idempRepo := setupCheckoutTestDB(t)
	uc := usecase.NewCheckoutUseCase(productRepo, orderRepo, idempRepo)
	ctx := context.Background()

	prod, err := domain.NewProduct("SKU-CH-1", "Gaming Mouse", "Wireless mouse", "Electronics", 50.0, 10, 0.2)
	require.NoError(t, err)
	require.NoError(t, productRepo.Create(ctx, prod))

	order, err := uc.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail: "buyer@example.com",
		Items: []usecase.CheckoutItemInput{
			{ProductID: prod.ID, Quantity: 2},
		},
		IdempotencyKey: "order-key-123",
	})
	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, 100.0, order.TotalAmount)
	assert.Equal(t, domain.OrderStatusCompleted, order.Status)
	assert.Len(t, order.Items, 1)

	updatedProd, err := productRepo.GetByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, 8, updatedProd.Stock)

	orderDup, err := uc.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail: "buyer@example.com",
		Items: []usecase.CheckoutItemInput{
			{ProductID: prod.ID, Quantity: 2},
		},
		IdempotencyKey: "order-key-123",
	})
	require.NoError(t, err)
	assert.Equal(t, order.ID, orderDup.ID)
	assert.Equal(t, order.OrderNumber, orderDup.OrderNumber)

	// Stock remains 8!
	updatedProd2, err := productRepo.GetByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, 8, updatedProd2.Stock)
}

func TestCheckoutUseCase_InsufficientStock(t *testing.T) {
	_, productRepo, orderRepo, idempRepo := setupCheckoutTestDB(t)
	uc := usecase.NewCheckoutUseCase(productRepo, orderRepo, idempRepo)
	ctx := context.Background()

	prod, err := domain.NewProduct("SKU-LOW-1", "Rare Item", "Limited edition", "Collectibles", 100.0, 1, 0.5)
	require.NoError(t, err)
	require.NoError(t, productRepo.Create(ctx, prod))

	// Requesting 2 when only 1 available
	order, err := uc.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail: "buyer@example.com",
		Items: []usecase.CheckoutItemInput{
			{ProductID: prod.ID, Quantity: 2},
		},
	})
	assert.ErrorIs(t, err, domain.ErrInsufficientStock)
	assert.Nil(t, order)

	// Stock remains untouched
	stillProd, err := productRepo.GetByID(ctx, prod.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, stillProd.Stock)
}

func TestCheckoutUseCase_ValidationErrors(t *testing.T) {
	_, productRepo, orderRepo, idempRepo := setupCheckoutTestDB(t)
	uc := usecase.NewCheckoutUseCase(productRepo, orderRepo, idempRepo)
	ctx := context.Background()

	// Invalid email
	_, err := uc.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail: "invalid-email",
		Items: []usecase.CheckoutItemInput{
			{ProductID: "some-id", Quantity: 1},
		},
	})
	assert.ErrorIs(t, err, domain.ErrInvalidCustomerEmail)

	// Empty items
	_, err = uc.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail: "valid@example.com",
		Items:         []usecase.CheckoutItemInput{},
	})
	assert.ErrorIs(t, err, usecase.ErrEmptyCheckoutItems)

	// Invalid quantity
	_, err = uc.Execute(ctx, usecase.CheckoutInput{
		CustomerEmail: "valid@example.com",
		Items: []usecase.CheckoutItemInput{
			{ProductID: "some-id", Quantity: 0},
		},
	})
	assert.ErrorIs(t, err, usecase.ErrInvalidItemQuantity)
}
