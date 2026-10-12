package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lucasgauffin/ecommerce-platform/internal/domain"
	"github.com/lucasgauffin/ecommerce-platform/internal/usecase"
)

type OrderHandler struct {
	checkoutUC *usecase.CheckoutUseCase
}

func NewOrderHandler(checkoutUC *usecase.CheckoutUseCase) *OrderHandler {
	return &OrderHandler{checkoutUC: checkoutUC}
}

type checkoutRequest struct {
	CustomerEmail  string                    `json:"customer_email"`
	Items          []usecase.CheckoutItemInput `json:"items"`
	IdempotencyKey string                    `json:"idempotency_key,omitempty"`
}

func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.IdempotencyKey == "" {
		idempHeader := r.Header.Get("Idempotency-Key")
		if idempHeader == "" {
			idempHeader = r.Header.Get("X-Idempotency-Key")
		}
		req.IdempotencyKey = strings.TrimSpace(idempHeader)
	}

	order, err := h.checkoutUC.Execute(r.Context(), usecase.CheckoutInput{
		CustomerEmail:  req.CustomerEmail,
		Items:          req.Items,
		IdempotencyKey: req.IdempotencyKey,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientStock):
			respondJSON(w, http.StatusConflict, map[string]string{
				"error": "insufficient stock for one or more items",
			})
		case errors.Is(err, domain.ErrInvalidCustomerEmail),
			errors.Is(err, domain.ErrEmptyOrderItems),
			errors.Is(err, usecase.ErrEmptyCheckoutItems),
			errors.Is(err, usecase.ErrInvalidItemQuantity),
			errors.Is(err, usecase.ErrEmptyProductID),
			errors.Is(err, domain.ErrInvalidStock):
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		case strings.Contains(err.Error(), "not found"):
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		default:
			respondJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal error processing order",
			})
		}
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "order ID is required",
		})
		return
	}

	order, err := h.checkoutUC.GetOrderByID(r.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, map[string]string{
				"error": "order not found",
			})
			return
		}
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal error fetching order",
		})
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

