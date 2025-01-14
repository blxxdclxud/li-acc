package repository

import (
	"context"
	"fmt"
	"li-acc/internal/model"
)

// SettingsRepository stores the object ot the DB Repository to manage the CRUD operations.
// Has following implemented methods: AddSetting, GetLastSetting
type SettingsRepository struct {
	db *Repository
}

// NewSettingsRepository creates and initializes new SettingsRepository object
func NewSettingsRepository(repo *Repository) *SettingsRepository {
	return &SettingsRepository{db: repo}
}

// GetLastSetting returns model.Settings object from the DB, stored as last record.
func (r *SettingsRepository) GetLastSetting(ctx context.Context) (model.Settings, error) {
	query := "SELECT ReceiptFile, Emails, QrPattern, SenderEmail FROM settings ORDER BY Id DESC LIMIT 1"

	rows, err := r.db.DB.Query(ctx, query) // get all rows from the table corresponding the query
	if err != nil {
		return model.Settings{}, fmt.Errorf("error during fetching settings: %w", err)
	}
	defer rows.Close()

	var setting model.Settings

	// Fill all fields of the settings model with fetched data
	err = rows.Scan(&setting.ReceiptFile, &setting.Emails, &setting.QrPattern, &setting.SenderEmail)
	if err != nil {
		return model.Settings{}, fmt.Errorf("error during scanning fetched setting: %w", err)
	}

	return setting, nil
}

// AddSetting adds new settings parameters to settings table in DB.
// settings parameter is the model containing all fields needed to store in table.
// New record stores ReceiptFile, Emails, QrPattern, SenderEmail.
func (r *SettingsRepository) AddSetting(ctx context.Context, settings model.Settings) error {

	query := "INSERT INTO settings (ReceiptFile, Emails, QrPattern, SenderEmail) VALUES ($1, $2, $3, $4)"

	_, err := r.db.DB.Exec(ctx, query, settings.ReceiptFile, settings.Emails, settings.QrPattern, settings.SenderEmail)
	if err != nil {
		return fmt.Errorf("error during inserting to settings table: %w", err)
	}
	return nil
}
