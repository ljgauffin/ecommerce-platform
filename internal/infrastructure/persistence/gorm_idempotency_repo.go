package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormIdempotencyRepo struct {
	db *gorm.DB
}

func NewGormIdempotencyRepo(db *gorm.DB) *GormIdempotencyRepo {
	return &GormIdempotencyRepo{db: db}
}

func (r *GormIdempotencyRepo) Get(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var rec domain.IdempotencyRecord
	err := r.db.WithContext(ctx).Where("key = ? AND expires_at > ?", key, time.Now().UTC()).First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *GormIdempotencyRepo) Save(ctx context.Context, record *domain.IdempotencyRecord) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(record).Error
}
