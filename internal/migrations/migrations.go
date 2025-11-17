// Package migrations provides migrations for the application.
package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hesoyamTM/nbf-auth/pkg/logger"
	_ "github.com/lib/pq"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	const op = "migrations.Migrate"
	log, err := logger.LoggerFromCtx(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Migrating database")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("No migrations to run")
			return nil
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Migrations completed")

	return nil
}
