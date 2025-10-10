//go:build integration

package integration

import (
	"context"
	"li-acc/internal/model"
	"li-acc/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSettingsRepository_AddSetting(t *testing.T) {
	ensureDBReady(t)

	want := []model.Settings{
		{
			ReceiptFile: []byte(`test receipt file 1`),
			Emails:      []byte(`{"test": "email"}`),
			QrPattern:   "test qr pattern",
			SenderEmail: "test-1@m.r",
		},
		{
			ReceiptFile: []byte(`test receipt file 2`),
			Emails:      []byte(`{"test": "email"}`),
			QrPattern:   "test qr pattern",
			SenderEmail: "test-2@m.r",
		},
	}

	s := repository.NewSettingsRepository(testRepo)

	for _, setting := range want {
		err := s.AddSetting(context.Background(), setting)
		require.NoError(t, err)

		time.Sleep(1 * time.Second)
	}

	query := "SELECT ReceiptFile, Emails, QrPattern, SenderEmail FROM settings"

	rows, err := testRepo.DB.Query(context.Background(), query) // get all rows from the table corresponding the query
	require.NoError(t, err)
	defer rows.Close()

	var settings []model.Settings

	for rows.Next() {
		var setting model.Settings

		// Fill all fields of the settings model with fetched data
		err = rows.Scan(&setting.ReceiptFile, &setting.Emails, &setting.QrPattern, &setting.SenderEmail)
		require.NoError(t, err)

		settings = append(settings, setting)
	}
	err = rows.Err()
	require.NoError(t, err)

	require.Equal(t, want, settings)
}

func TestSettingsRepository_GetLastSetting(t *testing.T) {
	ensureDBReady(t)

	in := []model.Settings{
		{
			ReceiptFile: []byte(`test receipt file firs`),
			Emails:      []byte(`{"test": "email"}`),
			QrPattern:   "first",
			SenderEmail: "test-1@m.r",
		},
		{
			ReceiptFile: []byte(`test receipt file 2`),
			Emails:      []byte(`{"test": "email"}`),
			QrPattern:   "last",
			SenderEmail: "test-2@m.r",
		},
	}

	s := repository.NewSettingsRepository(testRepo)

	// Add settings in the order as they come in `in`
	time.Sleep(1 * time.Second)
	for _, setting := range in {
		query := "INSERT INTO settings (ReceiptFile, Emails, QrPattern, SenderEmail) VALUES ($1, $2, $3, $4)"

		_, err := testRepo.DB.Exec(context.Background(), query,
			setting.ReceiptFile, setting.Emails, setting.QrPattern, setting.SenderEmail)
		require.NoError(t, err)

		time.Sleep(1 * time.Second)
	}

	// Get last setting
	last, err := s.GetLastSetting(context.Background())
	require.NoError(t, err)

	// Check that it is indeed last one
	require.Equal(t, in[len(in)-1], last)

}
