package sender

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormMessage(t *testing.T) {
	t.Run("without attachment", func(t *testing.T) {
		email, isAttach := FormMessage("Subject", "Body", "", "sender@test.com", "rec1@test.com", "rec2@test.com")
		require.Equal(t, []string{"sender@test.com"}, email.GetHeader("From"))
		require.Equal(t, []string{"rec1@test.com", "rec2@test.com"}, email.GetHeader("To"))
		require.Equal(t, isAttach, false)
	})

	t.Run("with valid attachment", func(t *testing.T) {
		// create temporary file
		tmpFile := filepath.Join(os.TempDir(), "test.txt")
		err := os.WriteFile(tmpFile, []byte("hello"), 0644)
		require.NoError(t, err)
		defer os.Remove(tmpFile)

		email, isAttach := FormMessage("Subject", "Body", tmpFile, "me@example.com", "you@example.com")
		require.Equal(t, []string{"me@example.com"}, email.GetHeader("From"))
		require.Equal(t, []string{"you@example.com"}, email.GetHeader("To"))
		require.Equal(t, isAttach, true)
	})

	t.Run("with not existing attachment", func(t *testing.T) {
		email, isAttach := FormMessage("Subject", "Body", "no-such-file.txt", "me@example.com", "you@example.com")
		require.Equal(t, isAttach, false)
		require.Equal(t, []string{"me@example.com"}, email.GetHeader("From"))
		require.Equal(t, []string{"you@example.com"}, email.GetHeader("To"))
	})
}
