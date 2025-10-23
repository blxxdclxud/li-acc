package mocks

import (
	"context"
	"li-acc/internal/model"

	"github.com/stretchr/testify/mock"
)

type HistoryService struct {
	mock.Mock
}

func (h *HistoryService) AddRecord(ctx context.Context, file model.File) error {
	args := h.Called(ctx, file)
	return args.Error(0)
}

func (h *HistoryService) GetRecords(ctx context.Context) ([]model.File, error) {
	args := h.Called(ctx)
	return args.Get(0).([]model.File), args.Error(1)
}
