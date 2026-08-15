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

// OrderCreatedEvent là payload publish lên Kafka topic "order-created".
// Schema khớp với system-architecture.md mục 3.
type OrderCreatedEvent struct {
	EventID     uuid.UUID          `json:"event_id"`
	EventType   string             `json:"event_type"`
	OccurredAt  time.Time          `json:"occurred_at"`
	OrderID     uuid.UUID          `json:"order_id"`
	CustomerID  uuid.UUID          `json:"customer_id"`
	Items       []OrderCreatedItem `json:"items"`
	TotalAmount float64            `json:"total_amount"`
}

type OrderCreatedItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}

func NewOrderCreatedEvent(order Order) OrderCreatedEvent {
	items := make([]OrderCreatedItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderCreatedItem{ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice}
	}
	return OrderCreatedEvent{
		EventID:     uuid.New(),
		EventType:   "order.created",
		OccurredAt:  time.Now().UTC(),
		OrderID:     order.ID,
		CustomerID:  order.CustomerID,
		Items:       items,
		TotalAmount: order.TotalAmount,
	}
}
