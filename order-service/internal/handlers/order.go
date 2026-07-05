package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-service/internal/models"
)

type OrderHandler struct {
	pool *pgxpool.Pool
}

func NewOrderHandler(pool *pgxpool.Pool) *OrderHandler {
	return &OrderHandler{pool: pool}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total float64
	for _, item := range req.Items {
		total += float64(item.Quantity) * item.UnitPrice
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	order := models.Order{CustomerID: req.CustomerID, Status: "created", TotalAmount: total}
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (customer_id, status, total_amount)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		order.CustomerID, order.Status, order.TotalAmount,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	order.Items = make([]models.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		oi := models.OrderItem{OrderID: order.ID, ProductID: item.ProductID, Quantity: item.Quantity, UnitPrice: item.UnitPrice}
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, unit_price)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			oi.OrderID, oi.ProductID, oi.Quantity, oi.UnitPrice,
		).Scan(&oi.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order items"})
			return
		}
		order.Items = append(order.Items, oi)
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	ctx := c.Request.Context()
	order, err := h.fetchOrder(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	customerIDParam := c.Query("customer_id")
	if customerIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id is required"})
		return
	}
	customerID, err := uuid.Parse(customerIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer_id"})
		return
	}

	ctx := c.Request.Context()
	rows, err := h.pool.Query(ctx,
		`SELECT id FROM orders WHERE customer_id = $1 ORDER BY created_at DESC`,
		customerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
		return
	}
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list orders"})
			return
		}
		ids = append(ids, id)
	}
	rows.Close()

	orders := make([]models.Order, 0, len(ids))
	for _, id := range ids {
		order, err := h.fetchOrder(ctx, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch order"})
			return
		}
		orders = append(orders, *order)
	}

	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) fetchOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := h.pool.QueryRow(ctx,
		`SELECT id, customer_id, status, total_amount, created_at, updated_at
		 FROM orders WHERE id = $1`,
		id,
	).Scan(&order.ID, &order.CustomerID, &order.Status, &order.TotalAmount, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := h.pool.Query(ctx,
		`SELECT id, product_id, quantity, unit_price FROM order_items WHERE order_id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order.Items = make([]models.OrderItem, 0)
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}
