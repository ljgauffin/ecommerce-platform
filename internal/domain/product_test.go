package domain_test

import (
	"testing"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct_Success(t *testing.T) {
	p, err := domain.NewProduct("RS-001", "Running Shoes", "Lightweight", "Footwear", 89.99, 150, 0.35)
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.NotEmpty(t, p.ID)
	assert.Equal(t, "RS-001", p.SKU)
	assert.Equal(t, "Running Shoes", p.Name)
	assert.Equal(t, 89.99, p.Price)
	assert.Equal(t, 150, p.Stock)
	assert.Equal(t, 0.35, p.WeightKg)
}

func TestNewProduct_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		sku         string
		prodName    string
		price       float64
		stock       int
		weight      float64
		expectedErr error
	}{
		{"empty sku", "", "Shoes", 10.0, 5, 0.5, domain.ErrInvalidSKU},
		{"whitespace sku", "   ", "Shoes", 10.0, 5, 0.5, domain.ErrInvalidSKU},
		{"empty name", "SKU-1", "", 10.0, 5, 0.5, domain.ErrInvalidName},
		{"whitespace name", "SKU-1", "   ", 10.0, 5, 0.5, domain.ErrInvalidName},
		{"negative price", "SKU-1", "Shoes", -1.0, 5, 0.5, domain.ErrInvalidPrice},
		{"negative stock", "SKU-1", "Shoes", 10.0, -1, 0.5, domain.ErrInvalidStock},
		{"negative weight", "SKU-1", "Shoes", 10.0, 5, -0.1, domain.ErrInvalidWeight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := domain.NewProduct(tt.sku, tt.prodName, "", "Cat", tt.price, tt.stock, tt.weight)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Nil(t, p)
		})
	}
}

func TestProduct_StockOperations(t *testing.T) {
	p, err := domain.NewProduct("SKU-1", "Item", "", "Cat", 10.0, 5, 1.0)
	require.NoError(t, err)

	err = p.DeductStock(3)
	require.NoError(t, err)
	assert.Equal(t, 2, p.Stock)

	err = p.DeductStock(3)
	assert.ErrorIs(t, domain.ErrInsufficientStock, err)
	assert.Equal(t, 2, p.Stock)

	err = p.DeductStock(0)
	assert.ErrorIs(t, domain.ErrInvalidStock, err)

	err = p.AddStock(10)
	require.NoError(t, err)
	assert.Equal(t, 12, p.Stock)
}

func TestNewOrder_Calculations(t *testing.T) {
	items := []domain.OrderItem{
		{
			ProductID:   "prod-1",
			SKU:         "SKU-1",
			ProductName: "Shoes",
			UnitPrice:   50.0,
			Quantity:    2,
		},
		{
			ProductID:   "prod-2",
			SKU:         "SKU-2",
			ProductName: "Socks",
			UnitPrice:   10.0,
			Quantity:    3,
		},
	}

	order, err := domain.NewOrder("test@example.com", items)
	require.NoError(t, err)
	require.NotNil(t, order)

	assert.Equal(t, 130.0, order.TotalAmount)
	assert.Equal(t, domain.OrderStatusCompleted, order.Status)
	assert.NotEmpty(t, order.PaymentReference)
	assert.Len(t, order.Items, 2)
	assert.Equal(t, 100.0, order.Items[0].Subtotal)
	assert.Equal(t, 30.0, order.Items[1].Subtotal)
}
