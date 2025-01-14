package repository

import (
	"context"
	"fmt"
	"li-acc/internal/model"
	"time"
)

// HistoryRepository stores the object ot the DB Repository to manage the CRUD operations.
// Has following implemented methods: AddHistory
type HistoryRepository struct {
	db *Repository
}

// NewHistoryRepository creates and initializes new HistoryRepository object
func NewHistoryRepository(repo *Repository) *HistoryRepository {
	return &HistoryRepository{db: repo}
}

// AddHistory adds new file to history table in DB.
// New record stores filename, file binary representation and the datetime of the file was added (Now).
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
