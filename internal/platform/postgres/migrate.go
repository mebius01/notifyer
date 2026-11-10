package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"notifyer/migrations"
)

// Migrate applies all pending embedded migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	const op = "postgres.Migrate"

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("%s: new provider: %w", op, err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("%s: up: %w", op, err)
	}
	return nil
}
