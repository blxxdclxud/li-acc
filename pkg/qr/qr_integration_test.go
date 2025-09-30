//go:build integration

package qr

import (
	"flag"
	"image/jpeg"
	"li-acc/pkg/model"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var keepQr = flag.Bool("keep-qr", false, "keep generated QRs for manual inspection")

// общий хелпер, чтобы не дублировать код
func generateAndCheckQR(t *testing.T, org model.Organization, payer model.Payer, outPath string) {
	t.Helper()

	qr := NewQrPattern(org)
	qrStr := qr.GetPayersQrDataString(payer)

	err := qr.GenerateQRCode(qrStr, outPath)
	require.NoError(t, err)

	// check file exists and not empty
	fi, err := os.Stat(outPath)
	require.NoError(t, err)
	require.Greater(t, fi.Size(), int64(10)) // size > 10 Bytes

	// check it's a valid JPEG
	f, err := os.Open(outPath)
	require.NoError(t, err)
	defer f.Close()

	_, err = jpeg.Decode(f)
	require.NoError(t, err)

	if !*keepQr {
		t.Cleanup(func() {
			_ = os.RemoveAll(filepath.Dir(outPath))
		})
	}
}

// существующий тест с фиктивными данными
func TestGenerateQRCode(t *testing.T) {
	qrPath := "./out/qr-integration-test.jpg"
	generateAndCheckQR(t, orgFix, payerFix, qrPath)
}

// новый тест с реальными реквизитами из .env.local
func TestGenerateQRCode_RealData(t *testing.T) {
	// грузим .env.local из текущей директории (qr/.env.local)
	if err := godotenv.Load(".env.local"); err != nil {
		t.Skip("no .env.local found, skipping real data test")
	}

	org := model.Organization{
		Name:        os.Getenv("NAME"),
		PersonalAcc: os.Getenv("PERSONALACC"),
		BankName:    os.Getenv("BANKNAME"),
		BIC:         os.Getenv("BIC"),
		CorrespAcc:  os.Getenv("CORRESPACC"),
		PayeeINN:    os.Getenv("PAYEEINN"),
		KPP:         os.Getenv("KPP"),
		ExtraParams: os.Getenv("EXTRAPARAMS"),
	}

	// если обязательных полей нет — пропускаем
	if org.Name == "" || org.PersonalAcc == "" || org.BIC == "" {
		t.Skip("real organization data not provided in .env.local")
	}

	// payer можно зашить тестовый (например, твой payerFix)
	payer := payerFix

	qrPath := "./out/qr-real-test.jpg"
	generateAndCheckQR(t, org, payer, qrPath)
}

// 5. Edge case: GenerateQRCode with empty string
func TestGenerateQRCode_EmptyData(t *testing.T) {
	qr := NewQrPattern(orgFix)
	tmpDir := t.TempDir()
	err := qr.GenerateQRCode("", tmpDir)
	require.Error(t, err)
}
