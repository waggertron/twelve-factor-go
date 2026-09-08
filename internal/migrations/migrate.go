package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 001_orders.sql
var migration001 string

func Apply(ctx context.Context, pool *pgxpool.Pool, target string) error {
	if target != "001" {
		return fmt.Errorf("unsupported migration target")
	}
	_, err := pool.Exec(ctx, migration001)
	return err
}

func Ready(ctx context.Context, pool *pgxpool.Pool) error {
	var present bool
	err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version='001')`).Scan(&present)
	if err != nil || !present {
		return fmt.Errorf("migration 001 is required")
	}
	return nil
}
