//go:build integration

package integration

import (
	"flag"
	"li-acc/pkg/pdf"
	"os"
	"path/filepath"
	"testing"

	"li-acc/pkg/model"
)

var keepPDF = flag.Bool("keep-pdf", false, "keep generated PDFs for manual inspection")
var debugMode = flag.Bool("debug", false, "activate debug mode to see frames on PDF")

func testPath(name string) string {
	return filepath.Join("testdata", name)
}

func TestGeneratePersonalReceipt_success(t *testing.T) {
	// test payer data
	payer := model.Payer{
		PersAcc:  "123456",
		CHILDFIO: "Зубенко Михаил Петрович",
		Purpose:  "10a доп питание сент",
		CBC:      "82100000000000000131",
		OKTMO:    "98790098",
		Sum:      "10150.40",
	}

	// origin pattern (must be in в testdata)
	pdfTemplatePath := testPath("template.pdf")
	qrImgFPath := testPath("qr-code.jpg")
	fontPath := testPath("Arial.ttf")

	// check, that needed files exist
	if _, err := os.Stat(pdfTemplatePath); err != nil {
		t.Skipf("missing template file: %s", pdfTemplatePath)
	}
	if _, err := os.Stat(qrImgFPath); err != nil {
		t.Skipf("missing QR image: %s", qrImgFPath)
	}

	// where to store the result
	var outDir string
	// if keep-pdf flag is set true, then store it in dedicated directory
	if *keepPDF {
		outDir = "./testdata/out"
		_ = os.MkdirAll(outDir, 0o755)
	} else {
		outDir = t.TempDir()
	}
	pdfDst := filepath.Join(outDir, "receipt.pdf")

	// call testing function
	err := pdf.GeneratePersonalReceipt(
		pdfTemplatePath,
		pdfDst,
		qrImgFPath,
		fontPath,
		payer,
		*debugMode,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// check taht file is created and is not empty
	info, err := os.Stat(pdfDst)
	if err != nil {
		t.Fatalf("pdf file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("generated pdf is empty")
	}
}

func TestGeneratePersonalReceipt_missingQrFile(t *testing.T) {
	payer := model.Payer{
		PersAcc:  "123456",
		CHILDFIO: "TEST NAME Surname",
		Purpose:  "test",
		CBC:      "123454656543",
		OKTMO:    "9879098",
		Sum:      "10150.40",
	}

	pdfTemplatePath := testPath("template.pdf")
	missingQrPath := testPath("not-exists.jpg")
	fontPath := testPath("Arial.ttf")
	pdfDst := filepath.Join(t.TempDir(), "receipt.pdf")

	// проверяем, что шаблон есть
	if _, err := os.Stat(pdfTemplatePath); err != nil {
		t.Skipf("missing template file: %s", pdfTemplatePath)
	}

	err := pdf.GeneratePersonalReceipt(
		pdfTemplatePath,
		pdfDst,
		missingQrPath,
		fontPath,
		payer,
		*debugMode,
	)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := err.Error(); got != "" && !contains(got, "failed to read a QR Code image file") {
		t.Errorf("unexpected error: %v", err)
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (filepath.Base(s) == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
