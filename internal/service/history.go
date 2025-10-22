package service

import (
	"context"
	"errors"
	"fmt"
	"li-acc/internal/model"
	"li-acc/internal/repository"
	"li-acc/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type HistoryRepo interface {
	AddHistory(ctx context.Context, f model.File) error
	GetHistory(ctx context.Context) ([]model.File, error)
}

type HistoryService interface {
	AddRecord(ctx context.Context, file model.File) error
	GetRecords(ctx context.Context) ([]model.File, error)
}

type historyService struct {
	repo HistoryRepo
}

func NewHistoryService(repo *repository.HistoryRepository) HistoryService {
	return &historyService{repo: repo}
}

// AddRecord adds a new file record into the database through the repository.
// It validates input, logs every step, measures execution time, and ensures
// proper error wrapping for traceability.
func (s *historyService) AddRecord(ctx context.Context, file model.File) error {
	start := time.Now()

	// Input validation
	if file.FileName == "" {
		err := errors.New("file name cannot be empty")
		logger.Warn("AddRecord validation failed", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	if len(file.FileData) == 0 {
		err := errors.New("file data cannot be empty")
		logger.Warn("AddRecord validation failed", zap.Error(err))
		return fmt.Errorf("validation error: %w", err)
	}

	// Context cancellation check (optional early exit)
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("AddRecord aborted - context canceled", zap.Error(err))
		return fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Attempt to insert record via repository
	logger.Info("AddRecord started",
		zap.String("file_name", file.FileName),
		zap.String("file_size", fmt.Sprintf("%.2f KB", float64(len(file.FileData))/1024)),
	)

	if err := s.repo.AddHistory(ctx, file); err != nil {
		logger.Error("AddRecord failed during repository insert",
			zap.String("file_name", file.FileName),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add record: %w", err)
	}

	// Log success and duration
	duration := time.Since(start)
	logger.Info("AddRecord completed successfully",
		zap.String("file_name", file.FileName),
		zap.Duration("elapsed", duration),
	)

	return nil
}

// GetRecords gets all file records from the database through the repository.
// It validates logs every step, measures execution time, and ensures
// proper error wrapping for traceability.
func (s *historyService) GetRecords(ctx context.Context) ([]model.File, error) {
	start := time.Now()

	// Context cancellation check (optional early exit)
	select {
	case <-ctx.Done():
		err := ctx.Err()
		logger.Warn("GetRecords aborted - context canceled", zap.Error(err))
		return nil, fmt.Errorf("operation canceled: %w", err)
	default:
	}

	// Attempt to insert record via repository
	logger.Info("GetRecords started")

	rec, err := s.repo.GetHistory(ctx)
	if err != nil {
		logger.Error("GetRecords failed during repository fetch",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get records: %w", err)
	}

	// Log success and duration
	var fn []string // files' names
	for _, f := range rec {
		fn = append(fn, f.FileName)
	}

	duration := time.Since(start)
	logger.Info("GetRecords completed successfully",
		zap.String("file_names", fmt.Sprint(fn)),
		zap.Duration("elapsed", duration),
	)

	return rec, nil
}
