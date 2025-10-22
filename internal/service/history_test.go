package service

import (
	"context"
	"errors"
	"li-acc/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//
// ===== Mock repository =====
//

// mockHistoryRepo simulates repository behavior for testing.
type mockHistoryRepo struct {
	addErr error
	getErr error
	files  []model.File
}

func (m *mockHistoryRepo) AddHistory(ctx context.Context, f model.File) error {
	return m.addErr
}

func (m *mockHistoryRepo) GetHistory(ctx context.Context) ([]model.File, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.files, nil
}

//
// ===== Helper =====
//

func newFile(name string, data string) model.File {
	return model.File{
		FileName: name,
		FileData: []byte(data),
	}
}

//
// ===== Tests for AddRecord =====
//

func TestAddRecord_Success(t *testing.T) {
	repo := &mockHistoryRepo{}
	svc := &historyService{repo: repo}

	file := newFile("report.xlsx", "mockdata")

	err := svc.AddRecord(context.Background(), file)
	require.NoError(t, err)
}

func TestAddRecord_ValidationFails(t *testing.T) {
	repo := &mockHistoryRepo{}
	svc := &historyService{repo: repo}

	tests := []struct {
		name string
		file model.File
		want string
	}{
		{"empty filename", model.File{FileData: []byte("data")}, "file name cannot be empty"},
		{"empty data", model.File{FileName: "x.txt", FileData: nil}, "file data cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AddRecord(context.Background(), tt.file)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestAddRecord_RepositoryError(t *testing.T) {
	repo := &mockHistoryRepo{addErr: errors.New("db insert failed")}
	svc := &historyService{repo: repo}

	file := newFile("data.xlsx", "mock")

	err := svc.AddRecord(context.Background(), file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "db insert failed")
}

func TestAddRecord_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo := &mockHistoryRepo{}
	svc := &historyService{repo: repo}

	file := newFile("data.xlsx", "mock")

	err := svc.AddRecord(ctx, file)
	require.Error(t, err)
	require.Contains(t, err.Error(), "operation canceled")
}

//
// ===== Tests for GetRecords =====
//

func TestGetRecords_Success(t *testing.T) {
	files := []model.File{
		newFile("file1.pdf", "data1"),
		newFile("file2.pdf", "data2"),
	}
	repo := &mockHistoryRepo{files: files}
	svc := &historyService{repo: repo}

	result, err := svc.GetRecords(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "file1.pdf", result[0].FileName)
}

func TestGetRecords_RepositoryError(t *testing.T) {
	repo := &mockHistoryRepo{getErr: errors.New("fetch failed")}
	svc := &historyService{repo: repo}

	_, err := svc.GetRecords(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "fetch failed")
}

func TestGetRecords_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	cancel()

	repo := &mockHistoryRepo{}
	svc := &historyService{repo: repo}

	_, err := svc.GetRecords(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "operation canceled")
}
