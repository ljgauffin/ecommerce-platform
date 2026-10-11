package domain

import (
	"context"
	"time"
)

type ProductFilter struct {
	Query       string
	Category    string
	MinPrice    *float64
	MaxPrice    *float64
	InStockOnly bool
	Limit       int
	Offset      int
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	GetBySKU(ctx context.Context, sku string) (*Product, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ProductFilter) ([]Product, int64, error)
	BulkUpsert(ctx context.Context, products []Product) (created int, updated int, err error)
}

type OrderRepository interface {
	CreateWithLock(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
}

type IdempotencyStatus string

const (
	IdempotencyStatusProcessing IdempotencyStatus = "PROCESSING"
	IdempotencyStatusCompleted  IdempotencyStatus = "COMPLETED"
	IdempotencyStatusFailed     IdempotencyStatus = "FAILED"
)

type IdempotencyRecord struct {
	Key          string            `json:"key" gorm:"primaryKey;type:varchar(128)"`
	Status       IdempotencyStatus `json:"status" gorm:"not null;type:varchar(32)"`
	ResponseCode int               `json:"response_code"`
	ResponseBody []byte            `json:"response_body" gorm:"type:bytea"`
	CreatedAt    time.Time         `json:"created_at" gorm:"not null;autoCreateTime"`
	ExpiresAt    time.Time         `json:"expires_at" gorm:"not null;index"`
}

type IdempotencyRepository interface {
	Get(ctx context.Context, key string) (*IdempotencyRecord, error)
	Save(ctx context.Context, record *IdempotencyRecord) error
}
