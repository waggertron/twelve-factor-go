package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/waggertron/twelve-factor-go/internal/domain"
)

type Orders interface {
	Create(context.Context, string, domain.OrderInput) (domain.Order, bool, error)
	Get(context.Context, string) (domain.Order, error)
	Claim(context.Context, string) (bool, error)
	Complete(context.Context, string) (bool, error)
	MarkEnqueued(context.Context, string) error
}

type Postgres struct{ Pool *pgxpool.Pool }

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func (p Postgres) Create(ctx context.Context, key string, input domain.OrderInput) (domain.Order, bool, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, false, err
	}
	defer tx.Rollback(ctx)
	now := time.Now().UTC()
	var order domain.Order
	err = tx.QueryRow(ctx, `
		INSERT INTO orders(id,idempotency_key,customer_id,amount_cents,status,created_at,updated_at)
		VALUES(gen_random_uuid(),$1,$2,$3,'accepted',$4,$4)
		ON CONFLICT(idempotency_key) DO NOTHING
		RETURNING id::text,idempotency_key,customer_id,amount_cents,status,created_at,updated_at`,
		key, input.CustomerID, input.AmountCents, now,
	).Scan(&order.ID, &order.IdempotencyKey, &order.CustomerID, &order.AmountCents, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	created := err == nil
	if errors.Is(err, pgx.ErrNoRows) {
		order, err = getWith(ctx, tx, key, true)
		if err == nil && (order.CustomerID != input.CustomerID || order.AmountCents != input.AmountCents) {
			return domain.Order{}, false, domain.ErrConflict
		}
	}
	if err != nil {
		return domain.Order{}, false, err
	}
	if created {
		if _, err = tx.Exec(ctx, `INSERT INTO queue_intents(order_id) VALUES($1)`, order.ID); err != nil {
			return domain.Order{}, false, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Order{}, false, err
	}
	return order, created, nil
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func getWith(ctx context.Context, q rowQuerier, value string, byKey bool) (domain.Order, error) {
	column := "id::text"
	if byKey {
		column = "idempotency_key"
	}
	var order domain.Order
	err := q.QueryRow(ctx, `SELECT id::text,idempotency_key,customer_id,amount_cents,status,created_at,updated_at FROM orders WHERE `+column+`=$1`, value).
		Scan(&order.ID, &order.IdempotencyKey, &order.CustomerID, &order.AmountCents, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, domain.ErrNotFound
	}
	return order, err
}

func (p Postgres) Get(ctx context.Context, id string) (domain.Order, error) {
	return getWith(ctx, p.Pool, id, false)
}

func (p Postgres) Claim(ctx context.Context, id string) (bool, error) {
	result, err := p.Pool.Exec(ctx, `UPDATE orders SET status='processing',updated_at=now() WHERE id=$1 AND status='accepted'`, id)
	return result.RowsAffected() == 1, err
}

func (p Postgres) Complete(ctx context.Context, id string) (bool, error) {
	result, err := p.Pool.Exec(ctx, `UPDATE orders SET status='completed',updated_at=now() WHERE id=$1 AND status='processing'`, id)
	return result.RowsAffected() == 1, err
}

func (p Postgres) MarkEnqueued(ctx context.Context, id string) error {
	_, err := p.Pool.Exec(ctx, `UPDATE queue_intents SET enqueued_at=now() WHERE order_id=$1 AND enqueued_at IS NULL`, id)
	return err
}
