package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID          uuid.UUID   `json:"id"`
	CustomerID  uuid.UUID   `json:"customer_id"`
	Status      string      `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id,omitempty"`
	ProductID  uuid.UUID `json:"product_id"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
}

type CreateOrderRequest struct {
	CustomerID uuid.UUID                `json:"customer_id" binding:"required"`
	Items      []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
	UnitPrice float64   `json:"unit_price" binding:"required,gt=0"`
}
