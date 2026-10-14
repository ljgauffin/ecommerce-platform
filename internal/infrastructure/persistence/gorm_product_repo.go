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
	err := r.db.WithContext(ctx).Preload("Category").First(&p, "id = ?", id).Error
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
	err := r.db.WithContext(ctx).Preload("Category").First(&p, "sku = ?", sku).Error
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

	joinedCategories := false
	if strings.TrimSpace(filter.Query) != "" {
		tx = tx.Joins("LEFT JOIN categories ON categories.id = products.category_id")
		joinedCategories = true
		q := "%" + strings.ToLower(strings.TrimSpace(filter.Query)) + "%"
		tx = tx.Where("LOWER(products.name) LIKE ? OR LOWER(products.sku) LIKE ? OR LOWER(products.description) LIKE ? OR LOWER(categories.name) LIKE ?", q, q, q, q)
	}

	if strings.TrimSpace(filter.CategoryID) != "" {
		tx = tx.Where("products.category_id = ?", strings.TrimSpace(filter.CategoryID))
	} else if strings.TrimSpace(filter.Category) != "" {
		if !joinedCategories {
			tx = tx.Joins("LEFT JOIN categories ON categories.id = products.category_id")
		}
		tx = tx.Where("LOWER(categories.name) = LOWER(?)", strings.TrimSpace(filter.Category))
	}

	if filter.MinPrice != nil {
		tx = tx.Where("products.price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		tx = tx.Where("products.price <= ?", *filter.MaxPrice)
	}

	if filter.StockStatus == "in_stock" || filter.InStockOnly {
		tx = tx.Where("products.stock > 0")
	} else if filter.StockStatus == "low_stock" {
		tx = tx.Where("products.stock > 0 AND products.stock <= 5")
	} else if filter.StockStatus == "out_of_stock" {
		tx = tx.Where("products.stock <= 0")
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

	orderClause := "products.created_at desc"
	switch filter.SortBy {
	case "price_asc", "price-asc":
		orderClause = "products.price asc"
	case "price_desc", "price-desc":
		orderClause = "products.price desc"
	case "name", "name_asc", "name-asc":
		orderClause = "products.name asc"
	case "name_desc", "name-desc":
		orderClause = "products.name desc"
	case "sku_asc", "sku-asc":
		orderClause = "products.sku asc"
	case "sku_desc", "sku-desc":
		orderClause = "products.sku desc"
	case "stock_asc", "stock-asc":
		orderClause = "products.stock asc"
	case "stock_desc", "stock-desc":
		orderClause = "products.stock desc"
	case "newest":
		orderClause = "products.created_at desc"
	}

	var products []domain.Product
	if err := tx.Preload("Category").Order(orderClause).Limit(limit).Offset(offset).Find(&products).Error; err != nil {
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
			"name", "description", "category_id", "price", "stock", "weight_kg", "updated_at",
		}),
	}).CreateInBatches(products, 100)

	if res.Error != nil {
		return 0, 0, res.Error
	}

	updated := int(existingCount)
	created := len(products) - updated
	return created, updated, nil
}
