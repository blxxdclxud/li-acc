package qr

import (
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Integration test: GenerateQRCode
func TestGenerateQRCode(t *testing.T) {
	qr := NewQrPattern(orgFix)
	qrStr := qr.GetPayersQrDataString(payerFix)

	qrPath := "./tmp/qr-integration-test.jpg"

	err := qr.GenerateQRCode(qrStr, qrPath)
	require.NoError(t, err)

	// check file exists and not empty
	fi, err := os.Stat(qrPath)
	require.NoError(t, err)
	require.Greater(t, fi.Size(), int64(10)) // size > 10 Bytes

	// check it's a valid JPEG
	f, err := os.Open(qrPath)
	require.NoError(t, err)
	defer f.Close()

	_, err = jpeg.Decode(f)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	// cleanup
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Dir(qrPath))
	})
}

// 5. Edge case: GenerateQRCode with empty string
func TestGenerateQRCode_EmptyData(t *testing.T) {
	qr := NewQrPattern(orgFix)
	tmpDir := t.TempDir()
	err := qr.GenerateQRCode("", tmpDir)
	require.Error(t, err)
}
