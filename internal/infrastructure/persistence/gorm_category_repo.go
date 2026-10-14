package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"gorm.io/gorm"
)

type GormCategoryRepo struct {
	db *gorm.DB
}

func NewGormCategoryRepo(db *gorm.DB) *GormCategoryRepo {
	return &GormCategoryRepo{db: db}
}

func (r *GormCategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	err := r.db.WithContext(ctx).Create(c).Error
	if err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "idx_categories_name") {
			return domain.ErrCategoryDuplicate
		}
		return err
	}
	return nil
}

func (r *GormCategoryRepo) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	var c domain.Category
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *GormCategoryRepo) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	var c domain.Category
	err := r.db.WithContext(ctx).First(&c, "LOWER(name) = LOWER(?)", strings.TrimSpace(name)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *GormCategoryRepo) FindOrCreate(ctx context.Context, name string) (*domain.Category, error) {
	existing, err := r.GetByName(ctx, name)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, domain.ErrCategoryNotFound) {
		return nil, err
	}

	c, err := domain.NewCategory(name)
	if err != nil {
		return nil, err
	}

	if err := r.Create(ctx, c); err != nil {
		if errors.Is(err, domain.ErrCategoryDuplicate) {
			return r.GetByName(ctx, name)
		}
		return nil, err
	}

	return c, nil
}

func (r *GormCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.WithContext(ctx).Order("name asc").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *GormCategoryRepo) Delete(ctx context.Context, id string) error {
	cat, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	var productCount int64
	err = r.db.WithContext(ctx).Model(&domain.Product{}).
		Where("category_id = ? OR LOWER(category) = LOWER(?)", cat.ID, cat.Name).
		Count(&productCount).Error
	if err != nil {
		return err
	}
	if productCount > 0 {
		return domain.ErrCategoryHasProducts
	}

	res := r.db.WithContext(ctx).Delete(&domain.Category{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}
