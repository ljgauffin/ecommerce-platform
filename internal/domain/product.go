package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid"`
	SKU         string    `json:"sku" gorm:"uniqueIndex;not null;type:varchar(64)"`
	Name        string    `json:"name" gorm:"not null;type:varchar(255);index"`
	Description string    `json:"description" gorm:"type:text"`
	CategoryID  string    `json:"category_id" gorm:"not null;type:uuid;index"`
	Category    *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Price       float64   `json:"price" gorm:"not null;type:numeric(12,2);index"`
	Stock       int       `json:"stock" gorm:"not null;default:0"`
	WeightKg    float64   `json:"weight_kg" gorm:"not null;type:numeric(8,3)"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func NewProduct(sku, name, description, categoryID string, price float64, stock int, weightKg float64) (*Product, error) {
	p := &Product{
		ID:          uuid.NewString(),
		SKU:         strings.TrimSpace(sku),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		CategoryID:  strings.TrimSpace(categoryID),
		Price:       price,
		Stock:       stock,
		WeightKg:    weightKg,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Product) Validate() error {
	if strings.TrimSpace(p.SKU) == "" {
		return ErrInvalidSKU
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidName
	}
	if p.Price < 0 {
		return ErrInvalidPrice
	}
	if p.Stock < 0 {
		return ErrInvalidStock
	}
	if p.WeightKg < 0 {
		return ErrInvalidWeight
	}
	return nil
}

func (p *Product) DeductStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidStock
	}
	if p.Stock < quantity {
		return ErrInsufficientStock
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Product) AddStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidStock
	}
	p.Stock += quantity
	p.UpdatedAt = time.Now().UTC()
	return nil
}
