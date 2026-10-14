package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryDuplicate   = errors.New("category already exists")
	ErrCategoryEmptyName   = errors.New("category name cannot be empty")
	ErrCategoryHasProducts = errors.New("cannot delete category with existing products")
)

type Category struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null;type:varchar(128)"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

func NewCategory(name string) (*Category, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrCategoryEmptyName
	}
	return &Category{
		ID:        uuid.NewString(),
		Name:      trimmed,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
