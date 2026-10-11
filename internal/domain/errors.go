package domain

import "errors"

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrProductDuplicateSKU     = errors.New("product with this SKU already exists")
	ErrInvalidSKU             = errors.New("invalid product SKU")
	ErrInvalidName            = errors.New("invalid product name: cannot be empty")
	ErrInvalidPrice           = errors.New("invalid product price: must be greater than or equal to 0")
	ErrInvalidStock           = errors.New("invalid product stock: must be non-negative integer")
	ErrInvalidWeight          = errors.New("invalid product weight: must be positive number")
	ErrInsufficientStock      = errors.New("insufficient stock for requested product")
	ErrEmptyOrderItems        = errors.New("order must contain at least one item")
	ErrInvalidCustomerEmail   = errors.New("invalid customer email address")
	ErrIdempotencyKeyRequired = errors.New("idempotency-key header is required")
	ErrIdempotencyConflict   = errors.New("concurrent request with the same idempotency key in progress")
)
