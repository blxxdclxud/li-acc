package mocks

import (
	"context"
	"li-acc/internal/model"

	"github.com/stretchr/testify/mock"
)

type SettingsService struct {
	mock.Mock
}

func (m *SettingsService) UploadSettings(ctx context.Context, settings model.Settings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func (m *SettingsService) GetSettings(ctx context.Context) (model.Settings, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.Settings), args.Error(1)
}

func (m *SettingsService) UploadEmails(ctx context.Context, emails map[string]string) error {
	args := m.Called(ctx, emails)
	return args.Error(0)
}

func (m *SettingsService) ProcessEmailsFile(ctx context.Context, filename string, fileData []byte) error {
	args := m.Called(ctx, filename, fileData)
	return args.Error(0)
}

func (m *SettingsService) SetSenderEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *SettingsService) GetCache() model.Settings {
	args := m.Called()
	return args.Get(0).(model.Settings)
}
