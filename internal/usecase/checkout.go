package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
)

var (
	ErrEmptyCheckoutItems  = errors.New("checkout requires at least one item")
	ErrInvalidItemQuantity = errors.New("quantity must be greater than zero")
	ErrEmptyProductID      = errors.New("product_id is required")
)

type CheckoutItemInput struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CheckoutInput struct {
	CustomerEmail  string              `json:"customer_email"`
	Items          []CheckoutItemInput `json:"items"`
	IdempotencyKey string              `json:"idempotency_key,omitempty"`
}

type CheckoutUseCase struct {
	productRepo     domain.ProductRepository
	orderRepo       domain.OrderRepository
	idempotencyRepo domain.IdempotencyRepository
}

func NewCheckoutUseCase(
	productRepo domain.ProductRepository,
	orderRepo domain.OrderRepository,
	idempotencyRepo domain.IdempotencyRepository,
) *CheckoutUseCase {
	return &CheckoutUseCase{
		productRepo:     productRepo,
		orderRepo:       orderRepo,
		idempotencyRepo: idempotencyRepo,
	}
}

func (uc *CheckoutUseCase) Execute(ctx context.Context, input CheckoutInput) (*domain.Order, error) {
	email := strings.TrimSpace(input.CustomerEmail)
	if email == "" || !strings.Contains(email, "@") {
		return nil, domain.ErrInvalidCustomerEmail
	}

	if len(input.Items) == 0 {
		return nil, ErrEmptyCheckoutItems
	}

	idempKey := strings.TrimSpace(input.IdempotencyKey)
	if idempKey != "" && uc.idempotencyRepo != nil {
		existing, err := uc.idempotencyRepo.Get(ctx, idempKey)
		if err != nil {
			return nil, fmt.Errorf("check idempotency key: %w", err)
		}
		if existing != nil && existing.Status == domain.IdempotencyStatusCompleted {
			var cachedOrder domain.Order
			if err := json.Unmarshal(existing.ResponseBody, &cachedOrder); err == nil {
				return &cachedOrder, nil
			}
		}
	}

	orderItems := make([]domain.OrderItem, len(input.Items))
	for i, itm := range input.Items {
		if strings.TrimSpace(itm.ProductID) == "" {
			return nil, ErrEmptyProductID
		}

		if itm.Quantity <= 0 {
			return nil, ErrInvalidItemQuantity
		}

		prod, err := uc.productRepo.GetByID(ctx, itm.ProductID)
		if err != nil {
			if errors.Is(err, domain.ErrProductNotFound) {
				return nil, fmt.Errorf("product '%s' not found: %w", itm.ProductID, err)
			}
			return nil, fmt.Errorf("fetch product '%s': %w", itm.ProductID, err)
		}

		if prod.Stock < itm.Quantity {
			return nil, domain.ErrInsufficientStock
		}

		orderItems[i] = domain.OrderItem{
			ProductID:   prod.ID,
			SKU:         prod.SKU,
			ProductName: prod.Name,
			UnitPrice:   prod.Price,
			Quantity:    itm.Quantity,
		}
	}

	order, err := domain.NewOrder(email, orderItems)
	if err != nil {
		return nil, err
	}

	if err := uc.orderRepo.CreateWithLock(ctx, order); err != nil {
		return nil, err
	}

	if idempKey != "" && uc.idempotencyRepo != nil {
		bodyBytes, _ := json.Marshal(order)
		rec := &domain.IdempotencyRecord{
			Key:          idempKey,
			Status:       domain.IdempotencyStatusCompleted,
			ResponseCode: http.StatusCreated,
			ResponseBody: bodyBytes,
			ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		}
		_ = uc.idempotencyRepo.Save(ctx, rec)
	}

	return order, nil
}

func (uc *CheckoutUseCase) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	return uc.orderRepo.GetByID(ctx, id)
}
