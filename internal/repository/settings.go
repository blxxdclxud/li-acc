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

	row := r.db.DB.QueryRow(ctx, query) // get a single row from the table corresponding the query

	var setting model.Settings

	// Fill all fields of the settings model with fetched data
	err := row.Scan(&setting.ReceiptFile, &setting.Emails, &setting.QrPattern, &setting.SenderEmail)
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
