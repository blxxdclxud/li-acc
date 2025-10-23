package service

import (
	"context"
	"errors"
	"li-acc/internal/errs"
	"li-acc/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockEmailParser struct {
	result map[string]string
	err    error
}

func (m *mockEmailParser) ParseEmail(_ string) (map[string]string, error) {
	return m.result, m.err
}

// ---- Mock repository ----

type mockSettingsRepo struct {
	setSettingsErr     error
	setEmailsErr       error
	setSenderEmailErr  error
	getSettingsErr     error
	getSettingsResult  model.Settings
	lastSetSettingsArg model.Settings
	lastSetEmailsArg   map[string]string
	lastSetSenderEmail string
}

func (m *mockSettingsRepo) SetSettings(ctx context.Context, s model.Settings) error {
	m.lastSetSettingsArg = s
	return m.setSettingsErr
}
func (m *mockSettingsRepo) GetSettings(ctx context.Context) (model.Settings, error) {
	if m.getSettingsErr != nil {
		return model.Settings{}, m.getSettingsErr
	}
	return m.getSettingsResult, nil
}
func (m *mockSettingsRepo) SetEmails(ctx context.Context, emails map[string]string) error {
	m.lastSetEmailsArg = emails
	return m.setEmailsErr
}
func (m *mockSettingsRepo) SetSenderEmail(ctx context.Context, email string) error {
	m.lastSetSenderEmail = email
	return m.setSenderEmailErr
}

// ---- Tests ----

func TestUploadSettings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		svc := &settingsService{repo: repo}

		settings := model.Settings{
			SenderEmail: "x@y.com",
			Emails:      map[string]string{"a": "b"},
		}

		err := svc.UploadSettings(context.Background(), settings)
		require.NoError(t, err)
		require.Equal(t, settings, svc.cache)
	})

	t.Run("validation - missing sender", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		svc := &settingsService{repo: (repo)}

		err := svc.UploadSettings(context.Background(), model.Settings{})
		require.ErrorContains(t, err, "sender email")
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockSettingsRepo{setSettingsErr: errors.New("db fail")}
		svc := &settingsService{repo: repo}

		settings := model.Settings{SenderEmail: "x@y.com", Emails: map[string]string{"a": "b"}}
		err := svc.UploadSettings(context.Background(), settings)
		require.ErrorContains(t, err, "db fail")
	})
}

func TestGetSettings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockSettingsRepo{getSettingsResult: model.Settings{SenderEmail: "ok"}}
		svc := &settingsService{repo: repo}

		got, err := svc.GetSettings(context.Background())
		require.NoError(t, err)
		require.Equal(t, "ok", got.SenderEmail)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockSettingsRepo{getSettingsErr: errors.New("db fail")}
		svc := &settingsService{repo: repo}
		_, err := svc.GetSettings(context.Background())
		require.ErrorContains(t, err, "db fail")
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		svc := &settingsService{repo: &mockSettingsRepo{}}
		_, err := svc.GetSettings(ctx)
		require.ErrorContains(t, err, "operation canceled")
	})
}

func TestUploadEmails(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		svc := &settingsService{repo: repo}

		emails := map[string]string{"a": "b"}
		err := svc.UploadEmails(context.Background(), emails)
		require.NoError(t, err)
		require.Equal(t, emails, svc.cache.Emails)
	})

	t.Run("validation empty", func(t *testing.T) {
		svc := &settingsService{}
		err := svc.UploadEmails(context.Background(), nil)
		require.ErrorContains(t, err, "emails map cannot be empty")
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockSettingsRepo{setEmailsErr: errors.New("insert fail")}
		svc := &settingsService{repo: repo}

		emails := map[string]string{"a": "b"}
		err := svc.UploadEmails(context.Background(), emails)
		require.ErrorContains(t, err, "insert fail")
	})
}

func TestSetSenderEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		svc := &settingsService{repo: repo}

		err := svc.SetSenderEmail(context.Background(), "x@y.com")
		require.NoError(t, err)
		require.Equal(t, "x@y.com", svc.cache.SenderEmail)
	})

	t.Run("validation empty", func(t *testing.T) {
		svc := &settingsService{}
		err := svc.SetSenderEmail(context.Background(), "")
		require.ErrorContains(t, err, "sender email cannot be empty")
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockSettingsRepo{setSenderEmailErr: errors.New("update fail")}
		svc := &settingsService{repo: repo}
		err := svc.SetSenderEmail(context.Background(), "x@y.com")
		require.ErrorContains(t, err, "update fail")
	})
}

func TestProcessEmailsFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		parser := &mockEmailParser{result: map[string]string{"John Doe": "john@example.com"}}
		svc := &settingsService{
			repo:   repo,
			parser: parser,
		}

		err := svc.ProcessEmailsFile(context.Background(), "emails.xlsx", []byte("mockdata"))
		require.NoError(t, err)
	})

	t.Run("parse user error", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		parser := &mockEmailParser{err: errs.New(errs.User, "invalid xls format")}
		svc := &settingsService{
			repo:   repo,
			parser: parser,
		}

		err := svc.ProcessEmailsFile(context.Background(), "emails.xlsx", []byte("bad"))
		require.Error(t, err)
		require.True(t, errs.IsUserError(err))
	})

	t.Run("parse system error", func(t *testing.T) {
		repo := &mockSettingsRepo{}
		parser := &mockEmailParser{err: errors.New("I/O read error")}
		svc := &settingsService{
			repo:   repo,
			parser: parser,
		}

		err := svc.ProcessEmailsFile(context.Background(), "emails.xlsx", []byte("bad"))
		require.Error(t, err)
		// even system errors are wrapped as errs.User at top-level in current code
		require.True(t, errs.IsUserError(err))
	})

	t.Run("upload emails fails", func(t *testing.T) {
		repo := &mockSettingsRepo{setEmailsErr: errors.New("repo fail")}
		parser := &mockEmailParser{result: map[string]string{"A": "a@example.com"}}
		svc := &settingsService{
			repo:   repo,
			parser: parser,
		}

		err := svc.ProcessEmailsFile(context.Background(), "emails.xlsx", []byte("mockdata"))
		require.Error(t, err)
		require.ErrorContains(t, err, "repo fail")
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		repo := &mockSettingsRepo{}
		parser := &mockEmailParser{result: map[string]string{"A": "a@example.com"}}
		svc := &settingsService{
			repo:   repo,
			parser: parser,
		}

		err := svc.ProcessEmailsFile(ctx, "emails.xlsx", []byte("mockdata"))
		require.Error(t, err)
		require.ErrorContains(t, err, "operation canceled")
	})
}

func TestGetCache(t *testing.T) {
	cache := model.Settings{SenderEmail: "cached@ok", Emails: map[string]string{"x": "y"}}
	svc := &settingsService{cache: cache}
	got := svc.GetCache()
	require.Equal(t, cache, got)
}
