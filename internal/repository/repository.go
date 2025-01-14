package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"li-acc/pkg/logger"
)

// Repository stores DB's connection pool object to manage operations with it.
type Repository struct {
	DB *pgxpool.Pool
}

// ConnectDB creates new connection pool and checks its connection using Ping.
// If it is succeeded, returns initialized Repository object.
func ConnectDB(dsn string) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("failed to connect to database", zap.String("dsn", dsn), zap.Error(err))
		return nil, err
	}

	// Check the connection to the DB
	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	logger.Info("Connected to PostgreSQL successfully")

	return &Repository{DB: pool}, nil

}

// CloseDB closes the pool of the Repository (closes the connection to DB)
func (r *Repository) CloseDB() {
	r.DB.Close()
}
