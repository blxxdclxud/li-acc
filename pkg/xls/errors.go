package xls

import (
	"fmt"
	"li-acc/internal/errs"
)

// ----- User-facing errors (validation / missing data) -----

// MissingParamsError error raised when certain sheet does not contain some necessary parameters (e.g. SettingsSheet).
// Implement interface errs.CodedError
type MissingParamsError struct {
	Sheet   string   // name of the sheet that contains error
	Missing []string // list of missing required parameters
}

func (e *MissingParamsError) Error() string {
	return fmt.Sprintf("sheet %s missing required parameters: %v", e.Sheet, e.Missing)
}

func (e *MissingParamsError) Kind() errs.Kind {
	return errs.User
}

// Unwrap — для совместимости, если оборачиваешь.
func (e *MissingParamsError) Unwrap() error {
	return nil
}

// MissingSheetError error raised when Excel file does not contain required sheet
type MissingSheetError struct {
	Sheet string
}

func (e *MissingSheetError) Error() string {
	return fmt.Sprintf("spreadsheet does not contain sheet: %s", e.Sheet)
}

func (e *MissingSheetError) Kind() errs.Kind {
	return errs.User
}

func (e *MissingSheetError) Unwrap() error {
	return nil
}

type MissingPayersSheetColumns struct {
	Want int
	Have int
}

func (e *MissingPayersSheetColumns) Error() string {
	return fmt.Sprintf("missing columns on payers sheet: expected %d, actual %d", e.Want, e.Have)
}

func (e *MissingPayersSheetColumns) Kind() errs.Kind {
	return errs.User
}

func (e *MissingPayersSheetColumns) Unwrap() error {
	return nil
}

// ----- Emails sheet incomplete rows -----

type MissingEmailsError struct {
	MissingLines []int // lines numbers that have missing either FIO or email
}

func (e *MissingEmailsError) Error() string {
	return "invalid emails sheet, missing parameters on lines: " + fmt.Sprint(e.MissingLines)
}

func (e *MissingEmailsError) Kind() errs.Kind {
	return errs.User
}

func (e *MissingEmailsError) Unwrap() error {
	return nil
}
