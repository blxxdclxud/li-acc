package xls

import "os"

// CreateSpreadSheetFile gets filename and
// data (binary representation of the .xls file have to be written to the physical file) parameters.
func CreateSpreadSheetFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, os.ModeType)
	if err != nil {
		return err
	}

	return nil
}
