package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"li-acc/pkg/logger"
)

type Repository struct {
	DB *pgxpool.Pool
}

func ConnectDB(dsn string) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("failed to connect to database", zap.String("dsn", dsn), zap.Error(err))
		return nil, err
	}

	// Check if the DB connection is working by doing a simple query
	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	logger.Info("Connected to PostgreSQL successfully")

	return &Repository{DB: pool}, nil

}

func (r *Repository) CloseDB() {
	r.DB.Close()
}
