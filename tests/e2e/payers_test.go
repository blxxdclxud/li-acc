//go:build e2e

package e2e

import (
	"encoding/json"
	"li-acc/internal/handler"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadPayersFile(t *testing.T) {
	env, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name               string
		emailsFile         string
		payersFile         string
		expectedHTTPStatus int
		expectedPartial    bool
		expectedEmailCount int
		expectedRecipients []string
		checkResponse      func(t *testing.T, resp handler.PayersFileUploadResponse)
		checkErrorResponse func(t *testing.T, resp *http.Response)
		checkEmails        func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse)
	}{
		// ========== SUCCESS CASES ==========
		{
			name:               "Valid XLS file - full success",
			emailsFile:         "testdata/emails/valid_emails.xlsx",
			payersFile:         "testdata/payers/valid_payers.xlsm",
			expectedHTTPStatus: http.StatusOK,
			expectedPartial:    false,
			expectedEmailCount: 2,
			expectedRecipients: []string{"ex1@example.com", "ex2@example.com"},
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.Equal(t, "file processed successfully", resp.Message)
				assert.False(t, resp.PartialSuccess)
				assert.Equal(t, 2, resp.SentAmount, "Should have sent exactly 2 emails")
				assert.Empty(t, resp.FailedEmails)
				assert.Empty(t, resp.MissingPayers)
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				// Verify correct number of emails
				assert.Len(t, emails, 2, "Should have sent exactly 2 emails")

				// Verify each email has PDF attachment
				for i, email := range emails {
					attachments := countEmailAttachments(&email)
					assert.Equal(t, 1, attachments, "Email %d should have exactly 1 PDF attachment, but got %d", i+1, attachments)
				}

				// Verify specific recipients
				recipients := []string{"ex1@example.com", "Ramzestwo2@yandex.ru"}
				for _, recipient := range recipients {
					msg := findEmailByRecipient(emails, recipient)
					assert.NotNil(t, msg, "Email should be sent to %s", recipient)
				}
			},
		},

		// ========== PARTIAL SUCCESS - EMAIL SENDING ERRORS ==========
		{
			name:               "Some emails fail to send (invalid SMTP addresses)",
			emailsFile:         "testdata/emails/some_invalid_smtp.xlsx",
			payersFile:         "testdata/payers/valid_payers.xlsm",
			expectedHTTPStatus: http.StatusOK,
			expectedPartial:    true,
			expectedEmailCount: 1, // Only 1 valid email
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.True(t, resp.PartialSuccess, "Should be partial success")
				assert.Equal(t, 1, resp.SentAmount, "Only 1 email should succeed")
				assert.NotEmpty(t, resp.FailedEmails, "Should have failed emails")
				assert.Empty(t, resp.MissingPayers)

				// Verify failed email list is populated
				for _, email := range resp.FailedEmails {
					assert.NotEmpty(t, email)
				}
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				// Only valid emails should be in MailHog
				assert.Len(t, emails, 1, "Only 1 email should be successfully sent")

				// Verify it has attachment
				if len(emails) > 0 {
					attachments := countEmailAttachments(&emails[0])
					assert.Equal(t, 1, attachments, "Email should have PDF attachment")
				}
			},
		},
		{
			name:               "All emails fail to send (all invalid SMTP)",
			emailsFile:         "testdata/emails/all_invalid_smtp.xlsx",
			payersFile:         "testdata/payers/valid_payers.xlsm",
			expectedHTTPStatus: http.StatusOK,
			expectedPartial:    true,
			expectedEmailCount: 0,
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.True(t, resp.PartialSuccess)
				assert.Equal(t, 0, resp.SentAmount, "No emails should be sent")
				assert.NotEmpty(t, resp.FailedEmails, "All emails should be in failed list")
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				assert.Empty(t, emails, "No emails should be in MailHog")
			},
		},

		// ========== PARTIAL SUCCESS - EMAIL MAPPING ERRORS ==========
		{
			name:               "Some payers have no matching emails",
			emailsFile:         "testdata/emails/partial_emails.xlsx",
			payersFile:         "testdata/payers/five_payers.xlsm",
			expectedHTTPStatus: http.StatusOK,
			expectedPartial:    true,
			expectedEmailCount: 3,
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.True(t, resp.PartialSuccess)
				assert.Equal(t, 3, resp.SentAmount, "Only 3 should succeed")
				assert.Empty(t, resp.FailedEmails)
				assert.Len(t, resp.MissingPayers, 2, "2 payers should have no emails")
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				assert.Len(t, emails, 3, "Should send exactly 3 emails")

				// All should have attachments
				for i, email := range emails {
					attachments := countEmailAttachments(&email)
					assert.Equal(t, 1, attachments, "Email %d should have PDF", i+1)
				}
			},
		},
		{
			name:               "No emails uploaded yet - all payers missing",
			emailsFile:         "",
			payersFile:         "testdata/payers/valid_payers.xlsm",
			expectedHTTPStatus: http.StatusBadRequest,
			expectedPartial:    true,
			expectedEmailCount: 0,
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.False(t, resp.PartialSuccess)
				assert.Equal(t, 0, resp.SentAmount)
				assert.Empty(t, resp.FailedEmails)
				assert.Empty(t, resp.MissingPayers)
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				assert.Empty(t, emails, "No emails should be sent")
			},
		},

		// ========== MIXED ERRORS ==========
		{
			name:               "Mixed: some missing emails + some invalid SMTP",
			emailsFile:         "testdata/emails/mixed_partial_and_invalid.xlsx",
			payersFile:         "testdata/payers/five_payers.xlsm",
			expectedHTTPStatus: http.StatusOK,
			expectedPartial:    true,
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.True(t, resp.PartialSuccess)
				assert.NotEmpty(t, resp.FailedEmails, "Should have SMTP send failures")
				assert.NotEmpty(t, resp.MissingPayers, "Should have mapping failures")
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				// Number of emails in MailHog should match SentAmount
				assert.Len(t, emails, resp.SentAmount, "MailHog count should match sent amount")
			},
		},

		// ========== BAD REQUEST ERRORS (400) ==========
		{
			name:               "Invalid file extension - .txt",
			emailsFile:         "testdata/emails/valid_emails.xlsx",
			payersFile:         "testdata/payers/invalid_extension.txt",
			expectedHTTPStatus: http.StatusBadRequest,
			checkErrorResponse: func(t *testing.T, resp *http.Response) {
				var errResp map[string]string
				json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], "xls")
			},
		},
		{
			name:               "No file uploaded",
			emailsFile:         "testdata/emails/valid_emails.xlsx",
			payersFile:         "",
			expectedHTTPStatus: http.StatusBadRequest,
			checkErrorResponse: func(t *testing.T, resp *http.Response) {
				var errResp map[string]string
				json.NewDecoder(resp.Body).Decode(&errResp)
				assert.Contains(t, errResp["error"], "file")
			},
		},

		// ========== SERVER ERRORS (500) ==========
		{
			name:               "Missing required columns",
			emailsFile:         "testdata/emails/valid_emails.xlsx",
			payersFile:         "testdata/payers/missing_columns.xlsm",
			expectedHTTPStatus: http.StatusBadRequest,
			checkErrorResponse: func(t *testing.T, resp *http.Response) {
				var errResp map[string]string
				json.NewDecoder(resp.Body).Decode(&errResp)
				assert.NotEmpty(t, errResp["error"])
			},
		},

		// ========== EDGE CASES ==========
		{
			name:               "Large file - 1000 payers",
			emailsFile:         "testdata/emails/1000_emails.xlsx",
			payersFile:         "testdata/payers/1000_payers.xlsm",
			expectedHTTPStatus: http.StatusOK,
			expectedPartial:    false,
			expectedEmailCount: 1000,
			checkResponse: func(t *testing.T, resp handler.PayersFileUploadResponse) {
				assert.Equal(t, 1000, resp.SentAmount)
				assert.False(t, resp.PartialSuccess)
			},
			checkEmails: func(t *testing.T, emails []MailHogMessage, resp handler.PayersFileUploadResponse) {
				var test []string
				for _, e := range emails {
					test = append(test, e.To[0].Mailbox)
				}
				require.Len(t, test, 1000, "Should send 1000 emails")

				// Spot check: verify random emails have attachments
				checkIndices := []int{0, 100, 500, 999}
				for _, idx := range checkIndices {
					if idx < len(emails) {
						attachments := countEmailAttachments(&emails[idx])
						assert.Equal(t, 1, attachments, "Email %d should have PDF", idx)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean DB and MailHog before each test
			cleanupDB(t, env)
			clearMailHog(t, env)

			// Setup: Upload emails file first (if specified)
			if tt.emailsFile != "" {
				setupEmailsCustom(t, env, tt.emailsFile)
			}

			// Special case: no payers file
			if tt.payersFile == "" {
				resp := uploadPayersFileNoFile(t, env)
				require.Equal(t, tt.expectedHTTPStatus, resp.StatusCode)
				if tt.checkErrorResponse != nil {
					tt.checkErrorResponse(t, resp)
				}
				return
			}

			// Test payers upload - error cases
			if tt.expectedHTTPStatus != http.StatusOK {
				resp := uploadPayersFileRaw(t, env, tt.payersFile)
				require.Equal(t, tt.expectedHTTPStatus, resp.StatusCode)
				if tt.checkErrorResponse != nil {
					tt.checkErrorResponse(t, resp)
				}
				return
			}

			// Test payers upload - success cases
			resp := uploadPayersFile(t, env, tt.payersFile)

			// Verify HTTP response
			if tt.checkResponse != nil {
				tt.checkResponse(t, resp)
			}

			// Verify emails in MailHog
			emails := getMailHogEmails(t, env)
			if tt.checkEmails != nil {
				tt.checkEmails(t, emails, resp)
			}

			// Verify expected email count
			if tt.expectedEmailCount > 0 {
				assert.Len(t, emails, tt.expectedEmailCount,
					"Expected %d emails in MailHog", tt.expectedEmailCount)
			}

			// Verify expected recipients (if specified)
			if len(tt.expectedRecipients) > 0 {
				for _, recipient := range tt.expectedRecipients {
					msg := findEmailByRecipient(emails, recipient)
					assert.NotNil(t, msg, "Expected email to %s", recipient)
				}
			}

			// Verify MailHog count matches response
			assert.Equal(t, resp.SentAmount, len(emails),
				"MailHog email count should match response SentAmount")

			// Verify metrics
			metrics := getMetrics(t, env)
			assert.Contains(t, metrics, "liacc_file_processed_total")
			assert.Contains(t, metrics, `type="payers"`)
		})
	}
}
