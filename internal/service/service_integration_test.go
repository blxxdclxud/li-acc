// run with API_KEY=<converter_api_public_key> go test -tags=integration ./internal/service
//go:build integration

package service

import (
	"context"
	"errors"
	"li-acc/internal/model"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cleanup = true
const pdfFontPath = "./testdata/Arial.ttf"

var converterKey = os.Getenv("API_KEY")

// TestIntegration_ProcessPayersFile_FileOrchestration tests the UNIQUE orchestration logic
// of the service layer: file storage → parsing → PDF/QR generation → email mapping.
// It does NOT re-test database operations (tested at repo layer) or individual pkg functions.
func TestIntegration_ProcessPayersFile_FileOrchestration(t *testing.T) {
	if converterKey == "" {
		t.Skip("no API_KEY in environment found")
	}

	ctx := context.Background()

	// Setup output directory
	outDir := "testdata/out"
	err := os.MkdirAll(outDir, 0755)
	require.NoError(t, err)

	// Clean up after test
	t.Cleanup(func() {
		os.RemoveAll(outDir)
	})

	// Load REAL XLS file with actual payer data
	xlsFile := "testdata/real_case_valid.xlsx"
	data, err := os.ReadFile(xlsFile)
	require.NoError(t, err, "Test asset should exist")

	// MOCK: Settings service with predefined email mappings (already tested at repo layer)
	mockSettings := &mockSettingsService{
		settings: model.Settings{
			Emails: map[string]string{
				"иванов иван": "john@example.com",
				"петров петр": "jane@example.com",
			},
			SenderEmail: "sender@example.com",
		},
		getErr: nil,
	}

	// MOCK: History service (already tested at repository layer)
	mockHistory := &mockHistoryService{
		addErr: nil,
	}

	// MOCK: Mail service (already tested in pkg/sender)
	mockMail := &mockMailService{
		sentCount: 2,
		sendErr:   nil,
	}

	// Create Manager with REAL file operations, MOCK database/email
	m := &Manager{
		History:            mockHistory,
		Settings:           mockSettings,
		Mail:               mockMail,
		storage:            defaultFileStorage{}, // REAL file operations
		payerParser:        defaultPayerParser{}, // REAL XLS parsing
		orgParser:          defaultOrgParser{},   // REAL XLS parsing
		converterConfigKey: converterKey,
		pdfFontPath:        pdfFontPath,
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   "./testdata/blank_receipt_pattern.xls",
			ReceiptPatternsDir: outDir,
			PayersXlsDir:       outDir,
			SentReceiptsDir:    outDir,
			QrCodesDir:         outDir,
		},
	}

	// === ACT: Execute the orchestration workflow ===
	startTime := time.Now()
	receiptsMap, sentCount, err := m.ProcessPayersFile(ctx, "real_case_valid.xlsm", data)
	elapsed := time.Since(startTime)

	require.Equal(t, sentCount, mockMail.sentCount)

	// === ASSERT: Verify orchestration worked correctly ===

	// 1. Overall success
	require.NoError(t, err, "ProcessPayersFile should succeed")
	require.NotEmpty(t, receiptsMap, "Should generate at least one receipt")
	t.Logf("ProcessPayersFile completed in %v, generated %d receipts", elapsed, len(receiptsMap))

	// 2. Receipt files were actually created with proper structure
	for email, pdfPath := range receiptsMap {
		t.Logf("Verifying receipt for %s at %s", email, pdfPath)

		// File exists
		require.FileExists(t, pdfPath, "PDF should be created at %s", pdfPath)

		// PDF is in output directory structure
		assert.Contains(t, pdfPath, outDir, "PDF should be in output directory")

		// Path contains timestamp (orchestration creates timestamped subdirs)
		assert.Regexp(t, `\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}`, pdfPath,
			"Path should contain timestamp from createNowDir()")

		// PDF has content and valid format
		pdfData, err := os.ReadFile(pdfPath)
		require.NoError(t, err)
		assert.Greater(t, len(pdfData), 1000, "PDF should have substantial content")
		assert.True(t, startsWithPDFMagicBytes(pdfData), "Should be valid PDF format")
	}

	// 3. Verify uploaded file was stored with timestamp
	payerFiles, err := filepath.Glob(filepath.Join(outDir, "*_real_case_valid.xlsm"))
	require.NoError(t, err)
	assert.NotEmpty(t, payerFiles, "Uploaded file should be stored with timestamp prefix")
	if len(payerFiles) > 0 {
		storedData, err := os.ReadFile(payerFiles[0])
		require.NoError(t, err)
		assert.Equal(t, data, storedData, "Stored file should match uploaded data")
	}

	// 4. Verify timestamped directories were created (orchestration logic)
	receiptsDirs, err := os.ReadDir(outDir)
	require.NoError(t, err)

	hasTimestampedDir := false
	for _, entry := range receiptsDirs {
		if entry.IsDir() && isTimestampedDirName(entry.Name()) {
			hasTimestampedDir = true
			t.Logf("Found timestamped directory: %s", entry.Name())

			// Verify QR codes exist in timestamped directory
			qrFiles, _ := filepath.Glob(filepath.Join(outDir, entry.Name(), "*.jpg"))
			assert.NotEmpty(t, qrFiles, "QR codes should be generated in %s", entry.Name())

			for _, qrPath := range qrFiles {
				qrData, _ := os.ReadFile(qrPath)
				assert.Greater(t, len(qrData), 100, "QR code %s should have content", qrPath)
			}
		}
	}
	assert.True(t, hasTimestampedDir, "At least one timestamped directory should exist")

	// 5. Verify number of receipts matches email mappings
	assert.Equal(t, len(mockSettings.settings.Emails), len(receiptsMap),
		"Should generate one receipt per email in settings")
}

// TestIntegration_ProcessPayersFile_ValidationFailure tests orchestration stops
// correctly when validation fails (no files should be created)
func TestIntegration_ProcessPayersFile_ValidationFailure(t *testing.T) {
	ctx := context.Background()

	outDir := "./testdata/out_validation_fail"
	_ = os.MkdirAll(outDir, 0755)
	if cleanup {
		t.Cleanup(func() { os.RemoveAll(outDir) })
	}

	xlsFile := "./testdata/real_case_valid.xlsx"
	data, err := os.ReadFile(xlsFile)
	require.NoError(t, err)

	// MOCK: Settings with missing emails (validation should fail)
	mockSettings := &mockSettingsService{
		settings: model.Settings{
			Emails:      nil, // Missing emails - validation will fail
			SenderEmail: "sender@example.com",
		},
		getErr: nil,
	}

	m := &Manager{
		History:     &mockHistoryService{},
		Settings:    mockSettings,
		Mail:        &mockMailService{},
		storage:     defaultFileStorage{},
		payerParser: defaultPayerParser{},
		orgParser:   defaultOrgParser{},
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   "./testdata/blank_receipt_pattern.xls",
			ReceiptPatternsDir: outDir,
			PayersXlsDir:       outDir,
			SentReceiptsDir:    outDir,
			QrCodesDir:         outDir,
		},
	}

	// ACT: Should fail at validation
	_, _, err = m.ProcessPayersFile(ctx, "test.xlsm", data)

	// ASSERT: Error returned, no files created
	require.Error(t, err)
	assert.Contains(t, err.Error(), "emails file is not uploaded")

	// Verify no PDFs were created (orchestration stopped early)
	pdfFiles, _ := filepath.Glob(filepath.Join(outDir, "**/*.pdf"))
	assert.Empty(t, pdfFiles, "No PDFs should be created when validation fails")
}

// TestIntegration_ProcessPayersFile_HistoryFailure tests orchestration continues
// and returns error when history recording fails
func TestIntegration_ProcessPayersFile_HistoryFailure(t *testing.T) {
	ctx := context.Background()

	outDir := "./testdata/out_history_fail"
	_ = os.MkdirAll(outDir, 0755)
	t.Cleanup(func() { os.RemoveAll(outDir) })

	xlsFile := "./testdata/real_case_valid.xlsx"
	data, err := os.ReadFile(xlsFile)
	require.NoError(t, err)

	mockSettings := &mockSettingsService{
		settings: model.Settings{
			Emails: map[string]string{
				"john doe": "john@example.com",
			},
			SenderEmail: "sender@example.com",
		},
	}

	// MOCK: History that fails (simulate DB error)
	mockHistory := &mockHistoryService{
		addErr: assert.AnError,
	}

	m := &Manager{
		History:     mockHistory,
		Settings:    mockSettings,
		Mail:        &mockMailService{},
		storage:     defaultFileStorage{},
		payerParser: defaultPayerParser{},
		orgParser:   defaultOrgParser{},
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   "./testdata/blank_receipt_pattern.xls",
			ReceiptPatternsDir: outDir,
			PayersXlsDir:       outDir,
			SentReceiptsDir:    outDir,
			QrCodesDir:         outDir,
		},
	}

	// ACT: Should fail at history recording
	_, _, err = m.ProcessPayersFile(ctx, "test.xlsm", data)

	// ASSERT: Error propagated correctly
	require.Error(t, err)
	assert.Contains(t, err.Error(), "history.AddRecord")

	// File was stored before history failure (orchestration order)
	storedFiles, _ := filepath.Glob(filepath.Join(outDir, "*_test.xlsm"))
	assert.NotEmpty(t, storedFiles, "File should be stored before history fails")
}

// Helper: Check if data starts with PDF magic bytes
func startsWithPDFMagicBytes(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "%PDF"
}

// Helper: Check if directory name matches timestamp format
func isTimestampedDirName(name string) bool {
	_, err := time.Parse("2006-01-02_15-04-05", name)
	return err == nil
}

// TestIntegration_ProcessPayersFile_EmailMappingError
// ensures that when some payers have missing emails, the service
// returns a partial receipts map and EmailMappingError error.
func TestIntegration_ProcessPayersFile_EmailMappingError(t *testing.T) {
	if converterKey == "" {
		t.Skip("no API_KEY in environment found")
	}

	ctx := context.Background()

	outDir := "./testdata/out_email_mapping"
	_ = os.MkdirAll(outDir, 0755)
	t.Cleanup(func() { os.RemoveAll(outDir) })

	xlsFile := "./testdata/real_case_valid.xlsx"
	data, err := os.ReadFile(xlsFile)
	require.NoError(t, err)

	// MOCK settings — missing one email intentionally
	mockSettings := &mockSettingsService{
		settings: model.Settings{
			Emails: map[string]string{
				"иванов иван": "john@example.com",
				// "петров петр": missing -> triggers EmailMappingError
			},
			SenderEmail: "sender@example.com",
		},
	}

	mockMail := &mockMailService{sentCount: 1}

	m := &Manager{
		History:            &mockHistoryService{},
		Settings:           mockSettings,
		Mail:               mockMail,
		storage:            defaultFileStorage{},
		payerParser:        defaultPayerParser{},
		orgParser:          defaultOrgParser{},
		converterConfigKey: converterKey,
		pdfFontPath:        pdfFontPath,
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   "./testdata/blank_receipt_pattern.xls",
			ReceiptPatternsDir: outDir,
			PayersXlsDir:       outDir,
			SentReceiptsDir:    outDir,
			QrCodesDir:         outDir,
		},
	}

	receiptsMap, sentCount, err := m.ProcessPayersFile(ctx, "real_case_valid.xlsm", data)

	// Must have partial success
	require.Error(t, err)
	require.NotEmpty(t, receiptsMap)
	require.Equal(t, sentCount, mockMail.sentCount)

	var mappingErr *EmailMappingError
	ok := errors.As(err, &mappingErr)
	require.True(t, ok, "Error must be of type EmailMappingError")

	t.Logf("EmailMappingError: %+v", mappingErr.MapPayerReceipt)
	assert.Greater(t, mappingErr.FailedCount(), 0, "at least one payer should be missing an email")
	assert.Equal(t, len(mockSettings.settings.Emails)+mappingErr.FailedCount(), len(receiptsMap)+mappingErr.FailedCount(), "should match total payers in xls")
}

// TestIntegration_ProcessPayersFile_EmailSendingError
// simulates a mail sending failure by using mockMailService with sendErr != nil
func TestIntegration_ProcessPayersFile_EmailSendingError(t *testing.T) {
	if converterKey == "" {
		t.Skip("no API_KEY in environment found")
	}

	ctx := context.Background()

	outDir := "./testdata/out_mail_error"
	_ = os.MkdirAll(outDir, 0755)
	t.Cleanup(func() { os.RemoveAll(outDir) })

	xlsFile := "./testdata/real_case_valid.xlsx"
	data, err := os.ReadFile(xlsFile)
	require.NoError(t, err)

	// Valid settings — no missing emails
	mockSettings := &mockSettingsService{
		settings: model.Settings{
			Emails: map[string]string{
				"иванов иван": "john@example.com",
				"петров петр": "jane@example.com",
			},
			SenderEmail: "sender@example.com",
		},
	}

	// Mock mail service that fails
	failingMail := &mockMailService{
		sendErr: &EmailSendingError{
			MapReceiverCause: map[string]string{
				"john@example.com": "SMTP 550 error",
			},
		},
		sentCount: 0,
	}

	m := &Manager{
		History:            &mockHistoryService{},
		Settings:           mockSettings,
		Mail:               failingMail,
		storage:            defaultFileStorage{},
		payerParser:        defaultPayerParser{},
		orgParser:          defaultOrgParser{},
		converterConfigKey: converterKey,
		pdfFontPath:        pdfFontPath,
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   "./testdata/blank_receipt_pattern.xls",
			ReceiptPatternsDir: outDir,
			PayersXlsDir:       outDir,
			SentReceiptsDir:    outDir,
			QrCodesDir:         outDir,
		},
	}

	receiptsMap, sentCount, err := m.ProcessPayersFile(ctx, "real_case_valid.xlsm", data)

	// Expect EmailSendingError
	require.Error(t, err)
	require.Equal(t, sentCount, failingMail.sentCount)
	var sendErr *EmailSendingError
	ok := errors.As(err, &sendErr)
	require.True(t, ok, "Error must be of type EmailSendingError")
	assert.GreaterOrEqual(t, sendErr.FailedCount(), 1, "Should report at least one recipient failure")

	t.Logf("EmailSendingError: %v", sendErr.MapReceiverCause)
	assert.NotEmpty(t, receiptsMap, "Receipts should still be generated even if mail sending failed")
}

// TestIntegration_ProcessPayersFile_CompositeError
// verifies that both EmailMappingError and EmailSendingError can be returned together
func TestIntegration_ProcessPayersFile_CompositeError(t *testing.T) {
	if converterKey == "" {
		t.Skip("no API_KEY in environment found")
	}

	ctx := context.Background()

	outDir := "./testdata/out_composite_error"
	_ = os.MkdirAll(outDir, 0755)
	t.Cleanup(func() { os.RemoveAll(outDir) })

	xlsFile := "./testdata/real_case_valid.xlsx"
	data, err := os.ReadFile(xlsFile)
	require.NoError(t, err)

	// MOCK: Missing emails for some payers and failing mail mock
	mockSettings := &mockSettingsService{
		settings: model.Settings{
			Emails: map[string]string{
				"иванов иван": "john@example.com",
				// "петров петр" missing intentionally
			},
			SenderEmail: "sender@example.com",
		},
	}

	failingMail := &mockMailService{
		sendErr: &EmailSendingError{
			MapReceiverCause: map[string]string{
				"john@example.com": "SMTP 554 transaction failed",
			},
		},
	}

	m := &Manager{
		History:            &mockHistoryService{},
		Settings:           mockSettings,
		Mail:               failingMail,
		storage:            defaultFileStorage{},
		payerParser:        defaultPayerParser{},
		orgParser:          defaultOrgParser{},
		converterConfigKey: converterKey,
		pdfFontPath:        pdfFontPath,
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   "./testdata/blank_receipt_pattern.xls",
			ReceiptPatternsDir: outDir,
			PayersXlsDir:       outDir,
			SentReceiptsDir:    outDir,
			QrCodesDir:         outDir,
		},
	}

	receiptsMap, sentCount, err := m.ProcessPayersFile(ctx, "real_case_valid.xlsm", data)

	require.Error(t, err)
	require.Equal(t, sentCount, failingMail.sentCount)
	var composite *CompositeError
	ok := errors.As(err, &composite)
	require.True(t, ok, "Error must be of type CompositeError")

	ers := composite.Unwrap()
	var mappingErr *EmailMappingError
	var sendingErr *EmailSendingError
	assert.True(t, errors.As(ers[0], &mappingErr))
	assert.True(t, errors.As(ers[1], &sendingErr))

	t.Logf("CompositeError contains %d sub-errors", len(composite.Errors))
	assert.NotEmpty(t, receiptsMap)
}
