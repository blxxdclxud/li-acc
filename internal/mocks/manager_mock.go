package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// Manager mocks ProcessPayersFile for tests
type Manager struct {
	mock.Mock
}

func (m *Manager) ProcessPayersFile(ctx context.Context, filename string, data []byte) (map[string]string, int, error) {
	args := m.Called(ctx, filename, data)
	return args.Get(0).(map[string]string), args.Int(1), args.Error(2)
}
