package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

const schema = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS orders (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id   UUID NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'created',
    total_amount  NUMERIC(12,2) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders(id),
    product_id  UUID NOT NULL,
    quantity    INT NOT NULL,
    unit_price  NUMERIC(12,2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
`

func InitSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}
