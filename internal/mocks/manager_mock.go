package mocks

import (
	"context"
	"li-acc/internal/service"

	"github.com/stretchr/testify/mock"
)

// Manager mocks ProcessPayersFile for tests
type Manager struct {
	mock.Mock
}

func (m *Manager) MailService() service.MailService {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) SettingsService() service.SettingsService {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) HistoryService() service.HistoryService {
	//TODO implement me
	panic("implement me")
}

func (m *Manager) ProcessPayersFile(ctx context.Context, filename string, data []byte) (map[string]string, int, error) {
	args := m.Called(ctx, filename, data)
	return args.Get(0).(map[string]string), args.Int(1), args.Error(2)
}
