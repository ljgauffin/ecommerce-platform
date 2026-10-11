package persistence

import (
	"context"
	"errors"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormOrderRepo struct {
	db *gorm.DB
}

func NewGormOrderRepo(db *gorm.DB) *GormOrderRepo {
	return &GormOrderRepo{db: db}
}

func (r *GormOrderRepo) CreateWithLock(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			var prod domain.Product
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&prod, "id = ?", item.ProductID).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrProductNotFound
				}
				return err
			}

			if prod.Stock < item.Quantity {
				return domain.ErrInsufficientStock
			}

			newStock := prod.Stock - item.Quantity
			if err := tx.Model(&domain.Product{}).Where("id = ?", item.ProductID).Update("stock", newStock).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *GormOrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return &order, nil
}
