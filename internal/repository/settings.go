package repository

import (
	"context"
	"fmt"
	"li-acc/internal/model"
)

type SettingsRepository struct {
	db *Repository
}

func NewSettingsRepository(repo *Repository) *SettingsRepository {
	return &SettingsRepository{db: repo}
}

func (r *SettingsRepository) GetLastSetting(ctx context.Context) (model.Settings, error) {
	query := "SELECT ReceiptFile, Emails, QrPattern, SenderEmail FROM settings ORDER BY Id DESC LIMIT 1"
	rows, err := r.db.DB.Query(ctx, query)
	if err != nil {
		return model.Settings{}, fmt.Errorf("error during fetching settings: %w", err)
	}
	defer rows.Close()

	var setting model.Settings

	err = rows.Scan(&setting.ReceiptFile, &setting.Emails, &setting.QrPattern, &setting.SenderEmail)
	if err != nil {
		return model.Settings{}, fmt.Errorf("error during scanning fetched setting: %w", err)
	}

	return setting, nil
}

func (r *SettingsRepository) AddSetting(ctx context.Context, settings model.Settings) error {

	query := "INSERT INTO settings (ReceiptFile, Emails, QrPattern, SenderEmail) VALUES ($1, $2, $3, $4)"

	_, err := r.db.DB.Exec(ctx, query, settings.ReceiptFile, settings.Emails, settings.QrPattern, settings.SenderEmail)
	if err != nil {
		return fmt.Errorf("error during inserting to settings table: %w", err)
	}
	return nil
}
