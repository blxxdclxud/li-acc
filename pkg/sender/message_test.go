package sender

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormMessage(t *testing.T) {
	t.Run("without attachment", func(t *testing.T) {
		email := FormMessage("Subject", "Body", "", "sender@test.com", "rec1@test.com", "rec2@test.com")
		require.Equal(t, Success, email.Status)
		require.Empty(t, email.StatusMsg)
		require.Equal(t, []string{"sender@test.com"}, email.Msg.GetHeader("From"))
		require.Equal(t, []string{"rec1@test.com", "rec2@test.com"}, email.Msg.GetHeader("To"))
	})

	t.Run("with valid attachment", func(t *testing.T) {
		// create temporary file
		tmpFile := filepath.Join(os.TempDir(), "test.txt")
		err := os.WriteFile(tmpFile, []byte("hello"), 0644)
		require.NoError(t, err)
		defer os.Remove(tmpFile)

		email := FormMessage("Subject", "Body", tmpFile, "me@example.com", "you@example.com")
		require.Equal(t, Success, email.Status)
		require.Empty(t, email.StatusMsg)
	})

	t.Run("with not existing attachment", func(t *testing.T) {
		email := FormMessage("Subject", "Body", "no-such-file.txt", "me@example.com", "you@example.com")
		require.Equal(t, Info, email.Status)
		require.Contains(t, email.StatusMsg, "attachment not found")
	})
}
