package domain

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

type OrderItem struct {
	ID          string  `json:"id" gorm:"primaryKey;type:uuid"`
	OrderID     string  `json:"order_id" gorm:"not null;type:uuid;index"`
	ProductID   string  `json:"product_id" gorm:"not null;type:uuid;index"`
	SKU         string  `json:"sku" gorm:"not null;type:varchar(64)"`
	ProductName string  `json:"product_name" gorm:"not null;type:varchar(255)"`
	UnitPrice   float64 `json:"unit_price" gorm:"not null;type:numeric(12,2)"`
	Quantity    int     `json:"quantity" gorm:"not null"`
	Subtotal    float64 `json:"subtotal" gorm:"not null;type:numeric(12,2)"`
}

type Order struct {
	ID               string      `json:"id" gorm:"primaryKey;type:uuid"`
	OrderNumber      string      `json:"order_number" gorm:"uniqueIndex;not null;type:varchar(64)"`
	CustomerEmail    string      `json:"customer_email" gorm:"not null;type:varchar(255);index"`
	TotalAmount      float64     `json:"total_amount" gorm:"not null;type:numeric(12,2)"`
	Status           OrderStatus `json:"status" gorm:"not null;type:varchar(32);default:'COMPLETED'"`
	PaymentReference string      `json:"payment_reference" gorm:"not null;type:varchar(64)"`
	Items            []OrderItem `json:"items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;"`
	CreatedAt        time.Time   `json:"created_at" gorm:"not null;autoCreateTime"`
}

func NewOrder(customerEmail string, items []OrderItem) (*Order, error) {
	email := strings.TrimSpace(customerEmail)
	if email == "" || !strings.Contains(email, "@") {
		return nil, ErrInvalidCustomerEmail
	}

	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}

	orderID := uuid.NewString()
	var total float64
	processedItems := make([]OrderItem, len(items))

	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidStock
		}
		subtotal := item.UnitPrice * float64(item.Quantity)
		total += subtotal

		processedItems[i] = OrderItem{
			ID:          uuid.NewString(),
			OrderID:     orderID,
			ProductID:   item.ProductID,
			SKU:         item.SKU,
			ProductName: item.ProductName,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		}
	}

	orderNumber := fmt.Sprintf("ORD-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))
	paymentRef := fmt.Sprintf("PAY-SIM-%s", uuid.NewString()[:8])

	return &Order{
		ID:               orderID,
		OrderNumber:      orderNumber,
		CustomerEmail:    email,
		TotalAmount:      total,
		Status:           OrderStatusCompleted,
		PaymentReference: paymentRef,
		Items:            processedItems,
		CreatedAt:        time.Now().UTC(),
	}, nil
}
