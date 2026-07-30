package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusCreated       = "CREATED"
	OrderStatusPaymentPending = "PAYMENT_PENDING"
	OrderStatusPaid          = "PAID"
	OrderStatusFailed        = "FAILED"
	OrderStatusCancelled     = "CANCELLED"
)

type Order struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ProductName string     `json:"product_name"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	PaymentID   *uuid.UUID `json:"payment_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateOrderRequest struct {
	UserID  string  `json:"user_id" binding:"required,uuid"`
	Product string  `json:"product" binding:"required,min=1,max=200"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}
