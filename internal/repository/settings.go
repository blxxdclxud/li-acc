package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"li-acc/internal/model"
)

// SettingsRepository stores the object ot the DB Repository to manage the CRUD operations.
// Has following implemented methods: SetSettings, GetSettings
type SettingsRepository struct {
	db *Repository
}

// NewSettingsRepository creates and initializes new SettingsRepository object
func NewSettingsRepository(repo *Repository) *SettingsRepository {
	return &SettingsRepository{db: repo}
}

// GetSettings returns model.Settings object from the DB, stored as last record.
func (r *SettingsRepository) GetSettings(ctx context.Context) (model.Settings, error) {
	query := "SELECT Emails, SenderEmail FROM settings WHERE Id=1"

	row := r.db.DB.QueryRow(ctx, query) // get a single row from the table corresponding the query

	var setting model.Settings

	// Fill all fields of the settings model with fetched data
	err := row.Scan(&setting.EmailsJSON, &setting.SenderEmail)
	if err != nil {
		return model.Settings{}, fmt.Errorf("error during scanning fetched setting: %w", err)
	}

	if err := setting.AfterLoad(); err != nil {
		return model.Settings{}, fmt.Errorf("error during serring.AfterLoad(): %w", err)
	}

	return setting, nil
}

// SetSettings adds new settings parameters to settings table in DB.
// settings parameter is the model containing all fields needed to store in table.
// New record stores Emails, SenderEmail.
func (r *SettingsRepository) SetSettings(ctx context.Context, settings model.Settings) error {

	query := `
		INSERT INTO settings (Id, Emails, SenderEmail)
		VALUES (1, $1, $2)
		ON CONFLICT (Id) DO UPDATE
		SET
			Emails = EXCLUDED.Emails,
			SenderEmail = EXCLUDED.SenderEmail
`

	// Store emails map as json string
	if err := settings.BeforeSave(); err != nil {
		return fmt.Errorf("error during settings.BeforeSave(): %w", err)
	}

	_, err := r.db.DB.Exec(ctx, query, settings.EmailsJSON, settings.SenderEmail)
	if err != nil {
		return fmt.Errorf("error during inserting to settings table: %w", err)
	}
	return nil
}

// SetEmails updates the Emails field in the single settings record (Id=1).
func (r *SettingsRepository) SetEmails(ctx context.Context, emails map[string]string) error {
	// Convert map to JSON string
	data, err := json.Marshal(emails)
	if err != nil {
		return fmt.Errorf("marshal emails to JSON: %w", err)
	}

	query := `
		INSERT INTO settings (Id, Emails)
		VALUES (1, $1)
		ON CONFLICT (Id) DO UPDATE SET Emails = EXCLUDED.Emails
`
	_, err = r.db.DB.Exec(ctx, query, string(data))
	if err != nil {
		return fmt.Errorf("update settings emails: %w", err)
	}

	return nil
}

// SetSenderEmail updates the sender email in the single settings record (Id=1).
func (r *SettingsRepository) SetSenderEmail(ctx context.Context, email string) error {
	query := `
		INSERT INTO settings (Id, SenderEmail)
		VALUES (1, $1)
		ON CONFLICT (Id) DO UPDATE SET SenderEmail = EXCLUDED.SenderEmail
`
	_, err := r.db.DB.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("update sender email: %w", err)
	}

	return nil
}
