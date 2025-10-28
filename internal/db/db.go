package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gbrvmm/L0/internal/model"
)

type Repository struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Repository{Pool: pool}, nil
}

func (r *Repository) Close() {
	if r.Pool != nil {
		r.Pool.Close()
	}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.Pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS orders (
	order_uid TEXT PRIMARY KEY,
	data JSONB NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bad_messages (
	id BIGSERIAL PRIMARY KEY,
	data TEXT,
	error TEXT,
	received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
	`)
	return err
}

func (r *Repository) SaveOrder(ctx context.Context, o model.Order) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = r.Pool.Exec(ctx, `
INSERT INTO orders (order_uid, data)
VALUES ($1, $2::jsonb)
ON CONFLICT (order_uid) DO NOTHING;
	`, o.OrderUID, string(b))
	return err
}

func (r *Repository) SaveBadMessage(ctx context.Context, raw []byte, errMsg string) error {
	_, err := r.Pool.Exec(ctx, `
INSERT INTO bad_messages (data, error) VALUES ($1, $2);
	`, string(raw), errMsg)
	return err
}

func (r *Repository) LoadAllOrders(ctx context.Context) (map[string]model.Order, error) {
	rows, err := r.Pool.Query(ctx, `SELECT order_uid, data FROM orders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]model.Order)
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		var o model.Order
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("unmarshal order %s: %w", id, err)
		}
		m[id] = o
	}
	return m, rows.Err()
}
