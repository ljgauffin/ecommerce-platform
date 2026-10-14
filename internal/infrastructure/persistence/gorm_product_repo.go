package persistence

import (
	"context"
	"errors"
	"strings"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormProductRepo struct {
	db *gorm.DB
}

func NewGormProductRepo(db *gorm.DB) *GormProductRepo {
	return &GormProductRepo{db: db}
}

func (r *GormProductRepo) Create(ctx context.Context, p *domain.Product) error {
	err := r.db.WithContext(ctx).Create(p).Error
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "idx_products_sku") || strings.Contains(errStr, "UNIQUE constraint") {
			return domain.ErrProductDuplicateSKU
		}
		return err
	}
	return nil
}

func (r *GormProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *GormProductRepo) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.WithContext(ctx).First(&p, "sku = ?", sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *GormProductRepo) Update(ctx context.Context, p *domain.Product) error {
	res := r.db.WithContext(ctx).Save(p)
	if res.Error != nil {
		errStr := res.Error.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "idx_products_sku") || strings.Contains(errStr, "UNIQUE constraint") {
			return domain.ErrProductDuplicateSKU
		}
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *GormProductRepo) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&domain.Product{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *GormProductRepo) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int64, error) {
	tx := r.db.WithContext(ctx).Model(&domain.Product{})

	if strings.TrimSpace(filter.Query) != "" {
		q := "%" + strings.ToLower(strings.TrimSpace(filter.Query)) + "%"
		tx = tx.Where("LOWER(name) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?", q, q, q, q)
	}

	if strings.TrimSpace(filter.Category) != "" {
		tx = tx.Where("LOWER(category) = LOWER(?)", strings.TrimSpace(filter.Category))
	}

	if filter.MinPrice != nil {
		tx = tx.Where("price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		tx = tx.Where("price <= ?", *filter.MaxPrice)
	}

	if filter.InStockOnly {
		tx = tx.Where("stock > 0")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 1000 {
		limit = 1000
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	orderClause := "created_at desc"
	switch filter.SortBy {
	case "price_asc", "price-asc":
		orderClause = "price asc"
	case "price_desc", "price-desc":
		orderClause = "price desc"
	case "name", "name_asc", "name-asc":
		orderClause = "name asc"
	case "newest":
		orderClause = "created_at desc"
	}

	var products []domain.Product
	if err := tx.Order(orderClause).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *GormProductRepo) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).Model(&domain.Category{}).
		Where("name IS NOT NULL AND name != ''").
		Distinct().
		Order("name asc").
		Pluck("name", &categories).Error
	if err == nil && len(categories) > 0 {
		return categories, nil
	}

	err = r.db.WithContext(ctx).Model(&domain.Product{}).
		Where("category IS NOT NULL AND category != ''").
		Distinct().
		Order("category asc").
		Pluck("category", &categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *GormProductRepo) BulkUpsert(ctx context.Context, products []domain.Product) (int, int, error) {
	if len(products) == 0 {
		return 0, 0, nil
	}

	skus := make([]string, len(products))
	for i, p := range products {
		skus[i] = p.SKU
	}

	var existingCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Product{}).Where("sku IN ?", skus).Count(&existingCount).Error; err != nil {
		return 0, 0, err
	}

	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "sku"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "description", "category", "price", "stock", "weight_kg", "updated_at",
		}),
	}).CreateInBatches(products, 100)

	if res.Error != nil {
		return 0, 0, res.Error
	}

	updated := int(existingCount)
	created := len(products) - updated
	return created, updated, nil
}
