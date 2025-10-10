package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// ApplyMigrations runs available DB migration files from `migrationsDir` directory.
//
//	Returns true, if there is new migration(-s) have been applied, false otherwise.
func ApplyMigrations(db *pgxpool.Pool, migrationsDir string) (bool, error) {
	// get database driver
	driver, err := postgres.WithInstance(
		stdlib.OpenDBFromPool(db), // convert *pgxpool.Pool type to *sql.DB
		&postgres.Config{},
	)
	if err != nil {
		return false, fmt.Errorf("failed to initialize golang-migrate driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres",
		driver,
	)
	if err != nil {
		return false, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return false, fmt.Errorf("failed to apply migrations: %w", err)
	}
	if errors.Is(err, migrate.ErrNoChange) { // check if no new migrations are applied
		return false, nil
	}
	return true, nil
}
