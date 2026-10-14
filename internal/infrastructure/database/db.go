package database

import (
	"context"
	"fmt"
	"time"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	Gorm *gorm.DB
}

func NewGormDB(ctx context.Context, dsn string) (*DB, error) {
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := gormDB.AutoMigrate(
		&domain.Category{},
		&domain.Product{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.IdempotencyRecord{},
	); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("auto migrate models: %w", err)
	}

	var existingCats []string
	gormDB.Model(&domain.Product{}).Where("category != ''").Distinct().Pluck("category", &existingCats)
	for _, name := range existingCats {
		var found domain.Category
		if err := gormDB.Where("LOWER(name) = LOWER(?)", name).First(&found).Error; err != nil {
			if c, err := domain.NewCategory(name); err == nil {
				gormDB.Create(c)
			}
		}
	}

	return &DB{Gorm: gormDB}, nil
}

func (db *DB) Close() error {
	if db.Gorm != nil {
		sqlDB, err := db.Gorm.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
