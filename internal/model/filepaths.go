package model

import "os"

const (
	TmpDir = "./tmp"

	PayersXlsDir       = TmpDir + "/payers"
	EmailXlsDir        = TmpDir + "/emails"
	ReceiptPatternsDir = TmpDir + "/patterns"
	SentReceiptsDir    = TmpDir + "/sent"
	QrCodesDir         = TmpDir + "/qr"
)

var BlankReceiptPath = "./assets/excel/blank_receipt_pattern.xls"

const MigrationsDir = "./internal/repository/db/migrations"

// EnsureTmpDirectories creates all required temporary directories
func EnsureTmpDirectories() error {
	dirs := []string{
		PayersXlsDir,
		EmailXlsDir,
		ReceiptPatternsDir,
		SentReceiptsDir,
		QrCodesDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
