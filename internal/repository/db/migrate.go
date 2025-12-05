package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// ApplyMigrations runs available DB migration files from `migrationsDir` directory.
//
//	Returns true, if there is new migration(-s) have been applied, false otherwise.
func ApplyMigrations(db *sql.DB, migrationsDir string) (bool, error) {
	// get database driver
	driver, err := postgres.WithInstance(
		db,
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
	defer migrator.Close()

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return false, fmt.Errorf("failed to apply migrations: %w", err)
	}
	if errors.Is(err, migrate.ErrNoChange) { // check if no new migrations are applied
		return false, nil
	}
	return true, nil
}
