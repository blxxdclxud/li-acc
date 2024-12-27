package repository

import (
	"context"
	"fmt"
	"li-acc/internal/model"
	"time"
)

type HistoryRepository struct {
	db *Repository
}

func NewHistoryRepository(repo *Repository) *HistoryRepository {
	return &HistoryRepository{db: repo}
}

func (r *HistoryRepository) AddHistory(ctx context.Context, file model.File) error {
	_, err := r.db.DB.Exec(ctx,
		`
			INSERT INTO files (FileName, File, ModifiedDate) 
			VALUES ($1, $2, $3)
		`,
		file.FileName, file.FileData, time.Now().Format("02-01-2006 15:04:05"))
	if err != nil {
		return fmt.Errorf("error during inserting to files table: %w", err)
	}
	return nil
}
