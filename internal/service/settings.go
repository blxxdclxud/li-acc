package service

import (
	"context"
	"errors"
	"fmt"
	"li-acc/internal/errs"
	"li-acc/internal/model"
	"li-acc/internal/repository"
	"li-acc/pkg/logger"
	"li-acc/pkg/xls"
	"time"

	"go.uber.org/zap"
)

type EmailParser interface {
	ParseEmail(path string) (map[string]string, error)
}

type xlsEmailParser struct{}

func (p *xlsEmailParser) ParseEmail(path string) (map[string]string, error) {
	return xls.ParseEmail(path)
}

type SettingsRepo interface {
	SetSettings(ctx context.Context, s model.Settings) error
	GetSettings(ctx context.Context) (model.Settings, error)
	SetEmails(ctx context.Context, emails map[string]string) error
	SetSenderEmail(ctx context.Context, email string) error
}

type SettingsService interface {
	UploadSettings(ctx context.Context, settings model.Settings) error
	GetSettings(ctx context.Context) (model.Settings, error)

	UploadEmails(ctx context.Context, emails map[string]string) error
	ProcessEmailsFile(ctx context.Context, filename string, fileData []byte) error

	SetSenderEmail(ctx context.Context, email string) error

	GetCache() model.Settings
}

type settingsService struct {
	repo   SettingsRepo
	cache  model.Settings
	parser EmailParser
}

func NewSettingsService(repo *repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo, parser: &xlsEmailParser{}}
}

func (s *settingsService) UploadSettings(ctx context.Context, settings model.Settings) error {
	start := time.Now()

	// Input validation
	if settings.SenderEmail == "" {
		err := errors.New("sender email cannot be empty")
		logger.Warn("UploadSettings validation failed", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	if settings.Emails == nil {
		err := errors.New("emails map cannot be empty")
		logger.Warn("UploadSettings validation failed", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	// Context cancellation check (optional early exit)
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("UploadSettings aborted - context canceled", zap.Error(err))
		return fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Attempt to insert record via repository
	logger.Info("UploadSettings started",
		zap.String("sender_email", settings.SenderEmail),
		zap.Int("emails_count", len(settings.Emails)),
	)

	if err := s.repo.SetSettings(ctx, settings); err != nil {
		logger.Error("UploadSettings failed during repository insert",
			zap.Error(err),
		)
		return fmt.Errorf("failed to add record: %w", err)
	}

	s.cache = settings

	// Log success and duration
	duration := time.Since(start)
	logger.Info("UploadSettings completed successfully",
		zap.Duration("elapsed", duration),
	)

	return nil
}

func (s *settingsService) GetSettings(ctx context.Context) (model.Settings, error) {
	start := time.Now()

	// Context cancellation check (optional early exit)
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("GetSettings aborted - context canceled", zap.Error(err))
		return model.Settings{}, fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Attempt to insert record via repository
	logger.Info("GetSettings started")

	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		logger.Error("GetSettings failed during repository fetch",
			zap.Error(err),
		)
		return model.Settings{}, fmt.Errorf("failed to get settings: %w", err)
	}

	// Update cache after fetching
	s.cache = settings

	// Log success and duration
	duration := time.Since(start)
	logger.Info("GetSettings completed successfully",
		zap.String("sender_email", settings.SenderEmail),
		zap.Int("emails_count", len(settings.Emails)),
		zap.Duration("elapsed", duration),
	)

	return settings, nil
}

// ProcessEmailsFile handles the full workflow of uploading and parsing
// an Excel file containing payer emails, then storing them in the system.
// It performs validation, logging, timing, and structured error wrapping.
//
// Steps:
//  1. Save uploaded Excel file to a predefined directory.
//  2. Parse the file to extract payer emails (map[FIO -> email]).
//  3. Store parsed emails using UploadEmails().
//  4. Log each stage with timing and context-aware cancellation.
//
// Returns:
//   - errs.User: if the Excel file has invalid or missing data.
//   - errs.System: if internal or IO-related failures occur.
func (s *settingsService) ProcessEmailsFile(ctx context.Context, filename string, fileData []byte) error {
	start := time.Now()
	logger.Info("ProcessEmailsFile started",
		zap.String("filename", filename),
		zap.String("size", fmt.Sprintf("%.2f KB", float64(len(fileData))/1024)),
	)

	// Context cancellation check
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("ProcessEmailsFile aborted - context canceled", zap.Error(err))
		return errs.Wrap(errs.System, "operation canceled", err)
	default:
	}

	// Step 1: Store uploaded Excel file
	filePath, err := storeUploadedFile(filename, model.EmailXlsDir, fileData)
	if err != nil {
		logger.Error("Failed to store uploaded emails file",
			zap.String("filename", filename),
			zap.Error(err),
		)
		return errs.Wrap(errs.System, "failed to store uploaded emails file", err)
	}
	logger.Info("Uploaded emails file stored successfully", zap.String("path", filePath))

	// Step 2: Parse Excel file to extract payer emails
	emails, err := s.parser.ParseEmail(filePath)
	if err != nil {
		// Если парсинг не удался из-за пользовательской ошибки (невалидный формат)
		// то оборачиваем её как errs.User
		logger.Warn("Failed to parse emails file",
			zap.String("filename", filename),
			zap.Error(err),
		)

		var c errs.CodedError
		if errors.As(err, &c) {
			if c.Kind() == errs.User {
				return errs.Wrap(errs.User, "invalid or corrupted emails file", err)
			}
		}

		return errs.Wrap(errs.System, "invalid or corrupted emails file", err)
	}

	logger.Info("Emails parsed successfully",
		zap.String("filename", filename),
		zap.Int("emails_count", len(emails)),
	)

	// Step 3: Upload parsed emails into settings storage
	if err := s.UploadEmails(ctx, emails); err != nil {
		logger.Error("UploadEmails failed",
			zap.String("filename", filename),
			zap.Int("emails_count", len(emails)),
			zap.Error(err),
		)
		return errs.Wrap(errs.System, "failed to upload parsed emails", err)
	}

	duration := time.Since(start)
	logger.Info("ProcessEmailsFile completed successfully",
		zap.String("filename", filename),
		zap.Int("emails_count", len(emails)),
		zap.Duration("elapsed", duration),
	)

	return nil
}

func (s *settingsService) UploadEmails(ctx context.Context, emails map[string]string) error {
	start := time.Now()

	// Input validation
	if emails == nil {
		err := errors.New("emails map cannot be empty")
		logger.Warn("UploadEmails validation failed", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	// Context cancellation check (optional early exit)
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("UploadEmails aborted - context canceled", zap.Error(err))
		return fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Attempt to insert record via repository
	logger.Info("UploadEmails started",
		zap.Int("emails_count", len(emails)),
	)

	if err := s.repo.SetEmails(ctx, emails); err != nil {
		logger.Error("UploadEmails failed during repository insert",
			zap.Error(err),
		)
		return fmt.Errorf("failed to set new emails: %w", err)
	}

	s.cache.Emails = emails

	// Log success and duration
	duration := time.Since(start)
	logger.Info("UploadEmails completed successfully",
		zap.Duration("elapsed", duration),
	)

	return nil
}

func (s *settingsService) SetSenderEmail(ctx context.Context, email string) error {
	start := time.Now()

	// Input validation
	if email == "" {
		err := errors.New("sender email cannot be empty")
		logger.Warn("SetSenderEmail validation failed", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	// Context cancellation check (optional early exit)
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("SetSenderEmail aborted - context canceled", zap.Error(err))
		return fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Attempt to insert record via repository
	logger.Info("SetSenderEmail started",
		zap.String("sender_email", email),
	)

	if err := s.repo.SetSenderEmail(ctx, email); err != nil {
		logger.Error("SetSenderEmail failed during repository insert",
			zap.Error(err),
		)
		return fmt.Errorf("failed to set new sender email: %w", err)
	}

	s.cache.SenderEmail = email

	// Log success and duration
	duration := time.Since(start)
	logger.Info("SetSenderEmail completed successfully",
		zap.String("sender_email", email),
		zap.Duration("elapsed", duration),
	)

	return nil
}

func (s *settingsService) GetCache() model.Settings {
	return s.cache
}
