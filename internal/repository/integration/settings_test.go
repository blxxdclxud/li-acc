//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"li-acc/internal/model"
	"li-acc/internal/repository"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettingsRepository_SetAndGetSettings(t *testing.T) {
	ensureDBReady(t)

	repo := repository.NewSettingsRepository(testRepo)

	emails := map[string]string{
		"John Doe":   "john@example.com",
		"Jane Smith": "jane@example.com",
	}

	set := model.Settings{
		Emails:      emails,
		SenderEmail: "sender@test.com",
	}

	// Save the setting
	err := repo.SetSettings(context.Background(), set)
	require.NoError(t, err)

	// Retrieve
	got, err := repo.GetSettings(context.Background())
	require.NoError(t, err)

	require.Equal(t, set.SenderEmail, got.SenderEmail)
	require.Equal(t, set.Emails, got.Emails)
}

func TestSettingsRepository_SetEmails(t *testing.T) {
	ensureDBReady(t)

	repo := repository.NewSettingsRepository(testRepo)

	// Initial emails
	initialEmails := map[string]string{
		"A": "a@test.com",
	}
	err := repo.SetEmails(context.Background(), initialEmails)
	require.NoError(t, err)

	// Check stored value
	var gotJSON string
	err = testRepo.DB.QueryRow(context.Background(), `SELECT Emails FROM settings WHERE Id=1`).Scan(&gotJSON)
	require.NoError(t, err)

	var gotMap map[string]string
	err = json.Unmarshal([]byte(gotJSON), &gotMap)
	require.NoError(t, err)
	require.Equal(t, initialEmails, gotMap)

	// Update emails
	newEmails := map[string]string{
		"B": "b@test.com",
		"C": "c@test.com",
	}
	err = repo.SetEmails(context.Background(), newEmails)
	require.NoError(t, err)

	gotMap = map[string]string{}

	// Check updated
	err = testRepo.DB.QueryRow(context.Background(), `SELECT Emails FROM settings WHERE Id=1`).Scan(&gotJSON)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(gotJSON), &gotMap)
	require.NoError(t, err)
	require.Equal(t, newEmails, gotMap)
}

func TestSettingsRepository_SetSenderEmail(t *testing.T) {
	ensureDBReady(t)

	repo := repository.NewSettingsRepository(testRepo)

	// Initial sender
	err := repo.SetSenderEmail(context.Background(), "initial@test.com")
	require.NoError(t, err)

	var got string
	err = testRepo.DB.QueryRow(context.Background(), `SELECT SenderEmail FROM settings WHERE Id=1`).Scan(&got)
	require.NoError(t, err)
	require.Equal(t, "initial@test.com", got)

	// Update sender
	err = repo.SetSenderEmail(context.Background(), "updated@test.com")
	require.NoError(t, err)

	err = testRepo.DB.QueryRow(context.Background(), `SELECT SenderEmail FROM settings WHERE Id=1`).Scan(&got)
	require.NoError(t, err)
	require.Equal(t, "updated@test.com", got)
}
