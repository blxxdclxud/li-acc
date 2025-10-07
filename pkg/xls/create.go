package xls

import (
	"li-acc/internal/errs"
	"os"
)

// CreateSpreadSheetFile gets filename and
// data (binary representation of the .xls file have to be written to the physical file) parameters.
func CreateSpreadSheetFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, 0o644)
	if err != nil {
		return errs.WrapIOError("create spreadsheet file", filename, err)
	}

	return nil
}
