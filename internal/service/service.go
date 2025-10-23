package service

import (
	"context"
	"errors"
	"fmt"
	"li-acc/internal/errs"
	"li-acc/internal/model"
	"li-acc/internal/repository"
	"li-acc/pkg/converter"
	"li-acc/pkg/logger"
	pkg "li-acc/pkg/model"
	"li-acc/pkg/pdf"
	"li-acc/pkg/qr"
	"li-acc/pkg/xls"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ===== Dependency interfaces for testability =====

// FileStorage writes uploaded files to disk or another medium.
type FileStorage interface {
	Store(filename, dir string, data []byte) (string, error)
}

// PayerParser extracts payer data from XLS files.
type PayerParser interface {
	ParsePayers(path string) ([]pkg.Payer, error)
}

// OrgParser extracts organization data from XLS files.
type OrgParser interface {
	ParseSettings(path string) (*pkg.Organization, error)
}

type defaultFileStorage struct{}

func (defaultFileStorage) Store(filename, dir string, data []byte) (string, error) {
	return storeUploadedFile(filename, dir, data)
}

type defaultPayerParser struct{}

func (defaultPayerParser) ParsePayers(path string) ([]pkg.Payer, error) {
	return xls.ParsePayers(path)
}

type defaultOrgParser struct{}

func (defaultOrgParser) ParseSettings(path string) (*pkg.Organization, error) {
	return xls.ParseSettings(path)
}

// Manager is the orchestrator that coordinates the domain services (history/settings/mail/...)
type Manager struct {
	History  HistoryService
	Settings SettingsService
	Mail     MailService
	repo     *repository.Repository

	storage     FileStorage
	payerParser PayerParser
	orgParser   OrgParser

	converterConfigKey string

	pdfFontPath string

	dirs struct {
		BlankReceiptPath   string
		ReceiptPatternsDir string
		PayersXlsDir       string
		SentReceiptsDir    string
		QrCodesDir         string
	}
}

// NewManager constructor
func NewManager(dsn string, converterConfig string, smtp model.SMTP) (*Manager, error) {
	repo, err := repository.ConnectDB(dsn)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		History:            NewHistoryService(repository.NewHistoryRepository(repo)),
		Settings:           NewSettingsService(repository.NewSettingsRepository(repo)),
		Mail:               NewMailService(smtp),
		repo:               repo,
		converterConfigKey: converterConfig,
		pdfFontPath:        pdf.DefaultFontPath,
		dirs: struct {
			BlankReceiptPath   string
			ReceiptPatternsDir string
			PayersXlsDir       string
			SentReceiptsDir    string
			QrCodesDir         string
		}{
			BlankReceiptPath:   model.BlankReceiptPath,
			ReceiptPatternsDir: model.ReceiptPatternsDir,
			PayersXlsDir:       model.PayersXlsDir,
			SentReceiptsDir:    model.SentReceiptsDir,
			QrCodesDir:         model.QrCodesDir,
		},
	}

	// inject defaults
	m.storage = defaultFileStorage{}
	m.payerParser = defaultPayerParser{}
	m.orgParser = defaultOrgParser{}

	return m, nil
}

// ProcessPayersFile handles the uploaded xls/xlsx file bytes: stores the file, parses payers and settings,
// generates receipts PDF files, sends emails with receipts and returns mapping email->pdfpath.
// It performs validation, logs every stage and preserves error kinds from lower-level packages.
// Return a non-nil CompositeError containing one or both EmailSendingError and EmailMappingError, or regular error.
func (m *Manager) ProcessPayersFile(ctx context.Context, filename string, data []byte) (map[string]string, int, error) {
	start := time.Now()
	logger.Info("ProcessPayersFile started", zap.String("filename", filename))

	// validate settings exist and are OK
	if err := m.validateBeforeProcessFile(ctx); err != nil {
		logger.Warn("validation before processing failed", zap.Error(err))
		return nil, 0, errs.Wrap(errs.System, "validation before processing failed", err)
	}

	settings := m.Settings.GetCache()

	// store uploaded file
	filePath, err := m.storage.Store(filename, m.dirs.PayersXlsDir, data)
	if err != nil {
		logger.Error("failed to store uploaded file", zap.String("path", filePath), zap.Error(err))
		return nil, 0, errs.Wrap(errs.System, "failed to store uploaded file "+filePath, err)
	}
	logger.Info("stored uploaded file", zap.String("path", filePath))

	// parse payers list
	payers, err := m.payerParser.ParsePayers(filePath)
	if err != nil {
		logger.Error("failed to parse payers from xls", zap.String("path", filePath), zap.Error(err))
		// preserve original error kind if present, that is already errs.System or errs.User
		return nil, 0, err
	}

	// parse organization settings from the same uploaded file (or separate)
	org, err := m.orgParser.ParseSettings(filePath)
	if err != nil {
		logger.Error("failed to parse settings from xls", zap.String("path", filePath), zap.Error(err))
		// preserve original error kind if present, that is already errs.System or errs.User
		return nil, 0, err
	}

	// record file in history
	if err := m.History.AddRecord(ctx, model.File{FileName: filename, FileData: data}); err != nil {
		logger.Error("failed to add history record", zap.Error(err))
		// Wrap system error and return
		return nil, 0, errs.Wrap(errs.System, "history.AddRecord: %w", err)
	}

	// formPersonalReceipts may return EmailMappingError or system error
	receiptsMap, err := m.formPersonalReceipts(ctx, payers, *org)
	var missedEmailsErr *EmailMappingError
	errorsCollected := []error{}

	if err != nil {
		if errors.As(err, &missedEmailsErr) {
			errorsCollected = append(errorsCollected, missedEmailsErr)
		} else {
			logger.Error("failed to form personal receipts", zap.Error(err))
			return nil, 0, errs.Wrap(errs.System, "formPersonalReceipts: ", err)
		}
	}

	emailsMap := settings.Emails
	var emailsList []string
	for _, e := range emailsMap {
		emailsList = append(emailsList, e)
	}

	mails := model.Mail{
		Subject:         model.MailDefaultSubject,
		Body:            model.MailDefaultBody,
		To:              emailsList,
		From:            m.Mail.GetSenderEmail(),
		AttachmentPaths: receiptsMap,
	}

	// SendMails may return EmailSendingError or system error
	sentCount, err := m.Mail.SendMails(ctx, mails)
	var failedMailsErr *EmailSendingError

	if err != nil {
		if errors.As(err, &failedMailsErr) {
			errorsCollected = append(errorsCollected, failedMailsErr)
		} else {
			logger.Error("failed to send some emails", zap.Error(err))
			return nil, 0, errs.Wrap(errs.System, "MailService.SendMails()", err)
		}
	}

	if len(errorsCollected) > 0 {
		// Return composite error containing all partial errors
		return receiptsMap, sentCount, &CompositeError{Errors: errorsCollected}
	}

	// No errors found, full success
	logger.Info("ProcessPayersFile completed",
		zap.String("filename", filename),
		zap.Int("payers_count", len(payers)),
		zap.Int("mails_sent", sentCount),
		zap.Duration("elapsed", time.Since(start)),
	)

	return receiptsMap, sentCount, nil

}

// formPersonalReceipts generates PDF receipts for each payer and returns map of receiver email -> pdf path.
// It does NOT send the emails; sending is responsibility of Mail service.
// If there are missed emails for some payers, they are not included in the result map, but custom EmailMappingError returned also.
func (m *Manager) formPersonalReceipts(ctx context.Context, payers []pkg.Payer, org pkg.Organization) (map[string]string, error) {
	start := time.Now()
	logger.Info("formPersonalReceipts started", zap.Int("payers_count", len(payers)))

	templatePath, err := m.prepareReceiptTemplate(org)
	if err != nil {
		logger.Error("prepareReceiptTemplate failed", zap.Error(err))
		return nil, err // system error
	}

	receiptsDir, err := createNowDir(m.dirs.SentReceiptsDir)
	if err != nil {
		logger.Error("failed to create receipts dir", zap.Error(err))
		return nil, err // system error
	}

	qrDir, err := createNowDir(m.dirs.QrCodesDir)
	if err != nil {
		logger.Error("failed to create qr dir", zap.Error(err))
		return nil, err // system error
	}

	qrCreator := qr.NewQrPattern(org)
	receiptsMap := make(map[string]string)

	missedPayers := make(map[string]string)

	// iterate payers
	for _, payer := range payers {
		select {
		case <-ctx.Done():
			logger.Warn("formPersonalReceipts aborted: context canceled")
			return nil, ctx.Err()
		default:
		}

		payerFileName := strings.ReplaceAll(strings.TrimSpace(payer.CHILDFIO), " ", "_")
		if payerFileName == "" {
			logger.Warn("empty payer name, skipping", zap.Any("payer", payer))
			continue
		}

		// create canvas per-payer (pdf object wraps the template)
		canvas, err := pdf.NewCanvasFromTemplate(templatePath, m.pdfFontPath, false) // debugMode true if logger present
		if err != nil {
			logger.Error("failed to create canvas from template", zap.Error(err))
			return nil, err
		}

		// qr file path
		qrFile := filepath.Join(qrDir, payerFileName+".jpg")
		qrString := qrCreator.GetPayersQrDataString(payer)
		if err := qrCreator.GenerateQRCode(qrString, qrFile); err != nil {
			logger.Error("failed to generate qr", zap.String("qrFile", qrFile), zap.Error(err))
			return nil, err
		}

		qrImgBytes, err := os.ReadFile(qrFile)
		if err != nil {
			logger.Error("failed to read qr image", zap.String("qrFile", qrFile), zap.Error(err))
			return nil, err
		}

		if err := canvas.Fill(payer, qrImgBytes); err != nil {
			logger.Error("canvas.Fill error", zap.Error(err))
			return nil, err
		}

		pdfFile := filepath.Join(receiptsDir, payerFileName+".pdf")
		if err := canvas.Save(pdfFile); err != nil {
			logger.Error("failed to save pdf", zap.String("pdf", pdfFile), zap.Error(err))
			return nil, err
		}

		payerEmail := m.Settings.GetCache().Emails[strings.ToLower(strings.TrimSpace(payer.CHILDFIO))]
		if payerEmail == "" {
			missedPayers[payer.CHILDFIO] = pdfFile
		} else {
			receiptsMap[payerEmail] = pdfFile
		}
	}

	var missedErr error
	if len(missedPayers) > 0 {
		missedErr = &EmailMappingError{MapPayerReceipt: missedPayers}
		logger.Warn("formPersonalReceipts missed emails for some payers", zap.Error(err))
	}

	logger.Info("formPersonalReceipts completed", zap.Int("generated", len(receiptsMap)), zap.Duration("elapsed", time.Since(start)))
	return receiptsMap, missedErr
}

// prepareReceiptTemplate creates XLSX based template with organization params and converts it to PDF using converter.
func (m *Manager) prepareReceiptTemplate(org pkg.Organization) (string, error) {
	start := time.Now()
	logger.Info("prepareReceiptTemplate started")

	xlsPath, err := xls.FillOrganizationParamsInReceipt(m.dirs.BlankReceiptPath, m.dirs.ReceiptPatternsDir, org)
	if err != nil {
		logger.Error("FillOrganizationParamsInReceipt failed", zap.Error(err))
		return "", errs.Wrap(errs.System, "FillOrganizationParamsInReceipt failed", err)
	}

	_, xlsFilename := filepath.Split(xlsPath)
	xlsCleanFilename := strings.TrimSuffix(xlsFilename, filepath.Ext(xlsFilename))
	pdfPath := filepath.Join(m.dirs.ReceiptPatternsDir, xlsCleanFilename+".pdf")

	logger.Info("converting xls to pdf", zap.String("xls", xlsPath), zap.String("pdf", pdfPath))
	if err := converter.ExcelToPdf(xlsPath, pdfPath, m.converterConfigKey); err != nil {
		logger.Error("ExcelToPdf failed", zap.Error(err))
		return "", errs.Wrap(errs.System, "ExcelToPdf failed", err)
	}

	logger.Info("prepareReceiptTemplate completed", zap.Duration("elapsed", time.Since(start)))
	return pdfPath, nil
}

// validateBeforeProcessFile ensures settings exist and are usable prior to processing.
func (m *Manager) validateBeforeProcessFile(ctx context.Context) error {
	// if GetSettings succeeded, settings are fetched into cache of settings service
	settings, err := m.Settings.GetSettings(ctx)
	if err != nil {
		logger.Warn("Settings.GetSettings failed", zap.Error(err))
		return errs.Wrap(errs.System, "Settings.GetSettings failed", err)
	}
	if settings.Emails == nil {
		return errs.New(errs.User, "emails file is not uploaded")
	}
	if settings.SenderEmail == "" {
		return errs.New(errs.User, "sender email is not set")
	}
	return nil
}

// storeUploadedFile stores uploaded bytes to a timestamped file path (returns full path).
// It ensures directory exists and writes file contents.
// Returns path to saved file.
func storeUploadedFile(filename, dir string, data []byte) (string, error) {
	// use timestamped filename - safer to use sanitized name and a session id or UUID
	now := time.Now().Unix()
	targetDir := filepath.Join(dir, fmt.Sprintf("%d_%s", now, filename))
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directories for '%s': %w", targetDir, err)
	}
	if err := xls.CreateSpreadSheetFile(targetDir, data); err != nil {
		return "", fmt.Errorf("failed to store uploaded Excel file: %w", err)
	}
	return targetDir, nil
}

// createNowDir makes a new directory inside base `dirPath` using current timestamp and returns its path.
func createNowDir(dirPath string) (string, error) {
	dir := filepath.Join(dirPath, time.Now().Format("2006-01-02_15-04-05")+"/")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdirall %s: %w", dir, err)
	}
	return dir, nil
}
