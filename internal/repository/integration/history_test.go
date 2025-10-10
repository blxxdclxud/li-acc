//go:build integration

package integration

import (
	"context"
	"li-acc/internal/model"
	"li-acc/internal/repository"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHistoryRepository_AddHistory(t *testing.T) {
	ensureDBReady(t)

	file := model.File{
		FileName: "test-file",
		FileData: []byte(`test data in test file 1234`),
	}

	h := repository.NewHistoryRepository(testRepo)
	err := h.AddHistory(context.Background(), file)
	require.NoError(t, err)

	files, err := h.GetHistory(context.Background())
	require.NoError(t, err)
	require.Equal(t, file, files[0])
}

func TestHistoryRepository_GetHistory(t *testing.T) {
	ensureDBReady(t)

	in := []model.File{
		{
			FileName: "test-file-1",
			FileData: []byte(`test data in test file 1234`),
		},
		{
			FileName: "test-file-2",
			FileData: []byte(`test data in test file 1234`),
		},
		{
			FileName: "test-file-3",
			FileData: []byte(`test data in test file 1234`),
		},
	}

	h := repository.NewHistoryRepository(testRepo)

	// Add files in the order as they come in `in`
	time.Sleep(1 * time.Second)
	for _, w := range in {
		err := h.AddHistory(context.Background(), w)
		require.NoError(t, err)

		time.Sleep(1 * time.Second)
	}

	// Get all files
	files, err := h.GetHistory(context.Background())
	require.NoError(t, err)

	// Check that order is reversed
	slices.Reverse(in)

	for i, wantFile := range in {
		require.Equal(t, wantFile, files[i])
	}
}
