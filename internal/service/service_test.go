package service

import (
	"context"
	"errors"
	"li-acc/internal/errs"
	"li-acc/internal/model"
	pkg "li-acc/pkg/model"
	"testing"

	"github.com/stretchr/testify/require"
)

//
// ======== Mocks ========
//

type mockSettingsService struct {
	settings model.Settings
	getErr   error
}

func (m *mockSettingsService) UploadSettings(context.Context, model.Settings) error { return nil }
func (m *mockSettingsService) GetSettings(context.Context) (model.Settings, error) {
	return m.settings, m.getErr
}
func (m *mockSettingsService) UploadEmails(context.Context, map[string]string) error { return nil }
func (m *mockSettingsService) ProcessEmailsFile(context.Context, string, []byte) error {
	return nil
}
func (m *mockSettingsService) SetSenderEmail(context.Context, string) error { return nil }
func (m *mockSettingsService) GetCache() model.Settings                     { return m.settings }

type mockHistoryService struct{ addErr error }

func (m *mockHistoryService) AddRecord(ctx context.Context, f model.File) error { return m.addErr }
func (m *mockHistoryService) GetRecords(ctx context.Context) ([]model.File, error) {
	return nil, nil
}

type mockMailService struct {
	sentCount int
	sendErr   error
}

func (m *mockMailService) SendMails(ctx context.Context, mail model.Mail) (int, error) {
	return m.sentCount, m.sendErr
}

func (m *mockMailService) GetSenderEmail() string {
	return "sender@example.com"
}

type mockFileStorage struct {
	path string
	err  error
}

func (m *mockFileStorage) Store(filename, dir string, data []byte) (string, error) {
	return m.path, m.err
}

type mockPayerParser struct {
	payers []pkg.Payer
	err    error
}

func (m *mockPayerParser) ParsePayers(_ string) ([]pkg.Payer, error) {
	return m.payers, m.err
}

type mockOrgParser struct {
	org *pkg.Organization
	err error
}

func (m *mockOrgParser) ParseSettings(_ string) (*pkg.Organization, error) {
	return m.org, m.err
}

//
// ======== Tests ========
//

func TestValidateBeforeProcessFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockSettingsService{
			settings: model.Settings{Emails: map[string]string{"A": "a@b.com"}, SenderEmail: "x@y.com"},
		}
		m := &Manager{Settings: svc}
		require.NoError(t, m.validateBeforeProcessFile(context.Background()))
	})

	t.Run("settings fetch error", func(t *testing.T) {
		svc := &mockSettingsService{getErr: errors.New("db down")}
		m := &Manager{Settings: svc}
		err := m.validateBeforeProcessFile(context.Background())
		require.Error(t, err)
		require.True(t, errs.IsSystemError(err))
	})

	t.Run("emails missing", func(t *testing.T) {
		svc := &mockSettingsService{
			settings: model.Settings{Emails: nil, SenderEmail: "x@y.com"},
		}
		m := &Manager{Settings: svc}
		err := m.validateBeforeProcessFile(context.Background())
		require.Error(t, err)
		require.True(t, errs.IsUserError(err))
		require.Contains(t, err.Error(), "emails file")
	})

	t.Run("sender email missing", func(t *testing.T) {
		svc := &mockSettingsService{
			settings: model.Settings{Emails: map[string]string{"a": "b"}, SenderEmail: ""},
		}
		m := &Manager{Settings: svc}
		err := m.validateBeforeProcessFile(context.Background())
		require.Error(t, err)
		require.True(t, errs.IsUserError(err))
		require.Contains(t, err.Error(), "sender email")
	})
}

func TestProcessPayersFile(t *testing.T) {
	ctx := context.Background()

	t.Run("validation fail", func(t *testing.T) {
		m := &Manager{
			Settings:    &mockSettingsService{getErr: errors.New("db fail")},
			storage:     &mockFileStorage{path: "mock.xlsx"},
			payerParser: &mockPayerParser{},
			orgParser:   &mockOrgParser{},
		}
		_, sentCount, err := m.ProcessPayersFile(ctx, "file.xlsx", []byte("data"))
		require.Error(t, err)
		require.True(t, errs.IsSystemError(err))
		require.Equal(t, sentCount, 0)
	})

	t.Run("store fail", func(t *testing.T) {
		m := &Manager{
			Settings:    &mockSettingsService{settings: model.Settings{Emails: map[string]string{"a": "b"}, SenderEmail: "c"}},
			storage:     &mockFileStorage{err: errors.New("io fail")},
			payerParser: &mockPayerParser{},
			orgParser:   &mockOrgParser{},
		}
		_, sentCount, err := m.ProcessPayersFile(ctx, "f.xlsx", []byte("x"))
		require.Error(t, err)
		require.True(t, errs.IsSystemError(err))
		require.Equal(t, sentCount, 0)
	})

	t.Run("parse payers fail", func(t *testing.T) {
		m := &Manager{
			Settings:    &mockSettingsService{settings: model.Settings{Emails: map[string]string{"a": "b"}, SenderEmail: "c"}},
			storage:     &mockFileStorage{path: "mock.xlsx"},
			payerParser: &mockPayerParser{err: errors.New("bad format")},
			orgParser:   &mockOrgParser{},
		}
		_, sentCount, err := m.ProcessPayersFile(ctx, "file.xlsx", []byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "bad format")
		require.Equal(t, sentCount, 0)
	})

	t.Run("parse org fail", func(t *testing.T) {
		m := &Manager{
			Settings:    &mockSettingsService{settings: model.Settings{Emails: map[string]string{"a": "b"}, SenderEmail: "c"}},
			storage:     &mockFileStorage{path: "mock.xlsx"},
			payerParser: &mockPayerParser{payers: []pkg.Payer{{CHILDFIO: "Jane"}}},
			orgParser:   &mockOrgParser{err: errors.New("org fail")},
		}
		_, sentCount, err := m.ProcessPayersFile(ctx, "file.xlsx", []byte("data"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "org fail")
		require.Equal(t, sentCount, 0)
	})

	t.Run("history fail", func(t *testing.T) {
		m := &Manager{
			History:     &mockHistoryService{addErr: errors.New("db write fail")},
			Settings:    &mockSettingsService{settings: model.Settings{Emails: map[string]string{"a": "b"}, SenderEmail: "c"}},
			storage:     &mockFileStorage{path: "mock.xlsx"},
			payerParser: &mockPayerParser{payers: []pkg.Payer{{CHILDFIO: "Jane"}}},
			orgParser:   &mockOrgParser{org: &pkg.Organization{Name: "Org"}},
		}
		_, sentCount, err := m.ProcessPayersFile(ctx, "f.xlsx", []byte("x"))
		require.Error(t, err)
		require.True(t, errs.IsSystemError(err))
		require.Equal(t, sentCount, 0)
	})

	t.Run("mail fail", func(t *testing.T) {
		m := &Manager{
			History:     &mockHistoryService{},
			Settings:    &mockSettingsService{settings: model.Settings{Emails: map[string]string{"a": "b"}, SenderEmail: "c"}},
			Mail:        &mockMailService{sendErr: errors.New("smtp fail")},
			storage:     &mockFileStorage{path: "mock.xlsx"},
			payerParser: &mockPayerParser{payers: []pkg.Payer{{CHILDFIO: "Jane"}}},
			orgParser:   &mockOrgParser{org: &pkg.Organization{Name: "Org"}},
		}
		_, sentCount, err := m.ProcessPayersFile(ctx, "f.xlsx", []byte("x"))
		require.Error(t, err)
		require.True(t, errs.IsSystemError(err))
		require.Equal(t, sentCount, 0)
	})
}
