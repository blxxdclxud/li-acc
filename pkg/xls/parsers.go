package xls

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"slices"
	"strings"
)

// constants that store sheet names
const (
	SettingsSheet = "Настройки"
	PayersSheet   = "Реестр начислений"
	EmailsSheet   = "emails"
)

// OpenAndCheckSheets gets filename of the xls file needed to open and the sheet name of the sheet, that's presence
// needs to be checked.
// Returns the excelize.File object to work with the loaded file, if no error occur.
func OpenAndCheckSheets(filename, sheetName string) (*excelize.File, error) {
	spreadsheet, err := excelize.OpenFile(filename) // load xls file
	if err != nil {
		return nil, err
	}

	// Check if the needed sheet is in the sheets list
	sheets := spreadsheet.GetSheetList()
	if !slices.Contains(sheets, SettingsSheet) {
		return nil, fmt.Errorf("spreadsheet does not contain sheet %s", sheetName)
	}
	return spreadsheet, nil
}

// ParseSettings parses all needed parameters from the sheet SettingsSheet in the given range (startRow, endRow).
// Parameters contain receiver information needed to perform payments.
func ParseSettings(filename string) (map[string]string, error) {
	var ss *excelize.File // spreadsheet

	// Check if the file is valid and open it
	ss, err := OpenAndCheckSheets(filename, SettingsSheet)
	if err != nil {
		return nil, err
	}

	// Get all rows of the sheet as 2d array of strings
	rows, err := ss.GetRows(SettingsSheet)
	if err != nil {
		return nil, err
	}

	// params will store all parsed parameters in the format `param_name: param_value`.
	// param_name is in cells in column A, and param_value in column B.
	params := make(map[string]string)

	startRow, endRow := 2, 12 // parameters are in the range of rows from 2 to 12 (A2:B12)

	for _, row := range rows[startRow-1 : endRow-1] { // A2:A12 and B2:B12
		if len(row) < 2 { // if there are missing either param_name or param_value, return an error
			return nil, fmt.Errorf("sheet %s is incomplete", SettingsSheet)
		}

		// write to the map
		params[row[0]] = row[1]
	}

	return params, nil
}

// ParsePayers parses rows for each payer from the sheet PayersSheet in the given range.
// Parsing starts from the given row (startRow).
// Rows contain payer and payment information as Name and Surname, bank number, payment amount, etc.
func ParsePayers(filename string) ([][]string, error) {
	var ss *excelize.File // spreadsheet

	// Check if the file is valid and open it
	ss, err := OpenAndCheckSheets(filename, SettingsSheet)
	if err != nil {
		return nil, err
	}

	// Get all rows of the sheet as 2d array of strings
	rows, err := ss.GetRows(PayersSheet)
	if err != nil {
		return nil, err
	}

	// parsed rows stored here
	var payers [][]string

	// from which row in the table need to start parsing
	startRow := 7

	for _, row := range rows[startRow-1:] {
		// if all cells in the row are empty, then ignore
		allEmpty := true
		for _, cell := range row {
			if cell != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		// a row with the payer data
		var payer []string

		for _, cell := range row {
			if strings.Contains(".", cell) { // check if the cell is the decimal number
				splitted := strings.Split(cell, ".")
				tail := splitted[len(splitted)-1]
				if len(tail) == 1 { // if there are only one decimal after the dot
					cell = strings.TrimSpace(cell) + "0" // Append "0" to ensure two decimal places
				}
			}
			payer = append(payer, strings.TrimSpace(cell))
		}

		payers = append(payers, payer)
	}
	return payers, nil
}

// ParseEmail parses all emails and payers full names from EmailsSheet.
// Each row contains two cells - Full Name and Email.
func ParseEmail(filename string) (map[string]string, error) {
	var ss *excelize.File // spreadsheet

	// Check if the file is valid and open it
	ss, err := OpenAndCheckSheets(filename, SettingsSheet)
	if err != nil {
		return nil, err
	}

	// Get all rows of the sheet as 2d array of strings
	rows, err := ss.GetRows(PayersSheet)
	if err != nil {
		return nil, err
	}

	// emails map will store emails in the format `full_name: email`
	emails := make(map[string]string)

	// start from row startRow in the table
	startRow := 2

	for _, row := range rows[startRow-1:] {
		if row[0] != "" && row[1] != "" { // if none of the two cells are empty
			emails[strings.TrimSpace(row[0])] = strings.TrimSpace(row[1]) // add the email to the map
		}
	}

	return emails, nil
}
