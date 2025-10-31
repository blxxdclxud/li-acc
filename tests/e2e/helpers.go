package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"li-acc/config"
	"li-acc/internal/handler"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const EndpointUploadPayers = "/api/upload-payers"
const EndpointUploadEmails = "/api/settings/upload-emails"

// uploadEmailsFileRaw uploads emails file and returns raw HTTP response (for testing errors)
func uploadEmailsFileRaw(t *testing.T, env *TestEnvironment, filePath string) *http.Response {
	t.Helper()

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest("POST", env.AppURL+EndpointUploadEmails, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// uploadEmailsFileNoFile tests uploading without a file
func uploadEmailsFileNoFile(t *testing.T, env *TestEnvironment) *http.Response {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest("POST", env.AppURL+EndpointUploadEmails, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// uploadEmailsFile uploads an emails configuration file
func uploadEmailsFile(t *testing.T, env *TestEnvironment, filePath string) {
	t.Helper()

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest("POST", env.AppURL+EndpointUploadEmails, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Email file upload should succeed")
}

// setupEmailsCustom uploads a custom emails file
func setupEmailsCustom(t *testing.T, env *TestEnvironment, emailsFile string) {
	t.Helper()
	uploadEmailsFile(t, env, emailsFile)
}

func uploadPayersFile(t *testing.T, env *TestEnvironment, filePath string) handler.PayersFileUploadResponse {
	t.Helper()

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest("POST", env.AppURL+EndpointUploadPayers, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result handler.PayersFileUploadResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

// uploadPayersFileRaw returns raw HTTP response for error checking
func uploadPayersFileRaw(t *testing.T, env *TestEnvironment, filePath string) *http.Response {
	t.Helper()

	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err)

	_, err = io.Copy(part, file)
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest("POST", env.AppURL+EndpointUploadPayers, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// uploadPayersFileNoFile uploads without file attached
func uploadPayersFileNoFile(t *testing.T, env *TestEnvironment) *http.Response {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest("POST", env.AppURL+EndpointUploadPayers, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// getMetrics fetches Prometheus metrics
func getMetrics(t *testing.T, env *TestEnvironment) string {
	t.Helper()

	resp, err := http.Get(env.MetricsURL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}

// cleanupDB truncates all test tables
func cleanupDB(t *testing.T, env *TestEnvironment) {
	t.Helper()

	// Create temporary connection just for cleanup
	cfg := config.LoadConfig("test")
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DbName)

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Truncate tables
	_, err = db.Exec("TRUNCATE TABLE settings, files CASCADE")
	require.NoError(t, err)

	// Restore sender email (same as NewManager sets)
	_, err = db.Exec(`
        INSERT INTO settings (Id, SenderEmail)
        VALUES (1, $1)
        ON CONFLICT (Id) DO UPDATE SET SenderEmail = EXCLUDED.SenderEmail
    `, cfg.SMTP.Email)
	require.NoError(t, err)

	t.Log("Database cleaned and sender email restored")
}
