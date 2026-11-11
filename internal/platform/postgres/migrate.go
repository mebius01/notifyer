package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"notifyer/migrations"
)

// Migrate applies all pending embedded migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	const op = "postgres.Migrate"

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("%s: new provider: %w", op, err)
	}
	migrations, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("%s: up: %w", op, err)
	}

	if len(migrations) == 0 {
		logger.InfoContext(ctx, "No pending migrations")
	}

	for _, m := range migrations {
		logger.InfoContext(ctx, "Migration applied",
			"version", m.Source.Version,
			"file", m.Source.Path,
			"duration", m.Duration)
	}
	return nil
}
