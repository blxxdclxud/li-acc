//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadEmailsFile(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name               string
		emailsFile         string
		expectedHTTPStatus int
		checkResponse      func(t *testing.T, body map[string]interface{})
	}{
		// ========== SUCCESS CASES ==========
		{
			name:               "Valid XLSX file - full success",
			emailsFile:         "testdata/emails/valid_emails.xlsx",
			expectedHTTPStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "file processed successfully", body["message"])
			},
		},

		// ========== VALIDATION ERRORS (400) ==========
		{
			name:               "Missing email addresses or payer names in some rows",
			emailsFile:         "testdata/emails/missing_emails.xlsx",
			expectedHTTPStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				errorMsg := body["error"].(string)
				assert.Contains(t, errorMsg, "пропущены данные")
				assert.Contains(t, errorMsg, "рядах", "Should mention row numbers")
			},
		},
		{
			name:               "Emails sheet is not present",
			emailsFile:         "testdata/emails/no_emails_sheet.xlsx",
			expectedHTTPStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				errorMsg := body["error"].(string)
				assert.Contains(t, errorMsg, "отсутствует", "Should mention empty file")
			},
		},

		// ========== BAD REQUEST ERRORS (400) ==========
		{
			name:               "Invalid file extension - .txt",
			emailsFile:         "testdata/emails/invalid_extension.txt",
			expectedHTTPStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				errorMsg := body["error"].(string)
				assert.Contains(t, errorMsg, "xls", "Should mention file type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean DB before each test
			cleanupDB(t, env)

			// Upload emails file
			resp := uploadEmailsFileRaw(t, env, tt.emailsFile)
			defer resp.Body.Close()

			// Verify HTTP status
			require.Equal(t, tt.expectedHTTPStatus, resp.StatusCode,
				"Expected status %d, got %d", tt.expectedHTTPStatus, resp.StatusCode)

			// Parse response body
			var body map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&body)
			require.NoError(t, err, "Failed to decode response body")

			// Run custom checks
			if tt.checkResponse != nil {
				tt.checkResponse(t, body)
			}

			// Verify metrics for successful uploads
			if tt.expectedHTTPStatus == http.StatusOK {
				metrics := getMetrics(t, env)
				assert.Contains(t, metrics, "liacc_file_processed_total")
				assert.Contains(t, metrics, `type="emails"`)
			}
		})
	}
}

// TestUploadEmailsFileNoFile tests the case when no file is uploaded
func TestUploadEmailsFileNoFile(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	cleanupDB(t, env)

	resp := uploadEmailsFileNoFile(t, env)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	errorMsg := body["error"].(string)
	assert.Contains(t, errorMsg, "file", "Should mention missing file")
}
