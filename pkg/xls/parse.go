package xls

import (
	"li-acc/internal/errs"
	"li-acc/pkg/model"
	"slices"
	"strings"

	"github.com/xuri/excelize/v2"
)

// constants that store sheet names
const (
	SettingsSheet = "Настройки"
	PayersSheet   = "Реестр начислений"
	EmailsSheet   = "emails"
)

// rows that are data borders on pages
const (
	// PayersRowStart from which row in the table need to start parsing
	PayersRowStart = 7

	// SettingsRowStart and SettingsRowEnd parameters are in the range of rows from 2 to 12 (A2:B12)
	SettingsRowStart = 2
	SettingsRowEnd   = 12

	// EmailsRowStart email mapping starts from row 2 in the table
	EmailsRowStart = 2
)

// OpenAndCheckSheets gets filename of the xls file needed to open and the sheet name of the sheet, that's presence
// needs to be checked.
// Returns the excelize.File object to work with the loaded file, if no error occur.
func OpenAndCheckSheets(filepath, sheetName string) (*excelize.File, error) {
	spreadsheet, err := excelize.OpenFile(filepath) // load xls file
	if err != nil {
		return nil, errs.WrapIOError("open spreadsheet", filepath, err)
	}

	// Check if the needed sheet is in the sheets list
	sheets := spreadsheet.GetSheetList()
	if !slices.Contains(sheets, sheetName) {
		spreadsheet.Close()
		return nil, &MissingSheetError{Sheet: sheetName}
	}
	return spreadsheet, nil
}

// get returns all rows of the specified Excel file sheet as [][]string.
// In case of a reading error (e.g., sheet not found), returns the system error errs.System.
func get(ss *excelize.File, sheet string) ([][]string, error) {
	rows, err := ss.GetRows(sheet)
	if err != nil {
		return nil, errs.Wrap(errs.System, "failed to get rows from excel", err)
	}
	return rows, nil
}

// ParseSettingsFromFile parses all needed parameters from the given sheet of settings in the given range (SettingsRowStart, SettingsRowEnd).
// Parameters contain receiver information needed to perform payments.
func ParseSettingsFromFile(ss *excelize.File, sheet string) (*model.Organization, error) {
	// Get all rows of the sheet as 2d array of strings
	rows, err := get(ss, sheet)
	if err != nil {
		return nil, err
	}

	// params will store all parsed parameters in the format `param_name: param_value`.
	// param_name is in cells in column A, and param_value in column B.
	params := make(map[string]string)

	for _, row := range rows[SettingsRowStart-1 : SettingsRowEnd] { // A2:A12 and B2:B12
		if len(row) < 2 { // if there are missing either param_name or param_value, just skip row
			continue
		}

		// write to the map
		key := strings.TrimSpace(row[0])
		val := strings.TrimSpace(row[1])
		if key != "" && val != "" {
			params[key] = val
		}
	}

	// requiredKeys contain parameters that must be present in the sheet
	requiredKeys := []string{
		"Наименование организации",
		"Расчетный счет",
		"Наименование банка",
		"БИК",
		"Корреспондентский счет",
		"ИНН",
		"КПП",
		"Дополнительные параметры ДШК",
	}

	// check if all required parameters exist in parsed params
	var missing []string
	for _, k := range requiredKeys {
		if _, ok := params[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 { // if any required parameters are missing, return an error
		return nil, &MissingParamsError{Missing: missing, Sheet: sheet}
	}

	orgData := model.Organization{
		Name:        params["Наименование организации"],
		PersonalAcc: params["Расчетный счет"],
		BankName:    params["Наименование банка"],
		BIC:         params["БИК"],
		CorrespAcc:  params["Корреспондентский счет"],
		PayeeINN:    params["ИНН"],
		KPP:         params["КПП"],
		ExtraParams: params["Дополнительные параметры ДШК"],
	}

	return &orgData, nil
}

// ParseSettings opens Excel file and checks its validity calling OpenAndCheckSheets.
// If no error, then calls ParseSettingsFromFile that performs parsing logic.
func ParseSettings(filepath string) (*model.Organization, error) {
	sheet := SettingsSheet // this is current settings sheet that must be in the spreadsheet
	ss, err := OpenAndCheckSheets(filepath, sheet)

	// Check if the file is valid and open it
	if err != nil {
		return nil, err
	}
	defer ss.Close()

	return ParseSettingsFromFile(ss, sheet)
}

// ParsePayersFromFile parses rows for each payer from the shee in the given range.
// Parsing starts from the given row (PayersRowStart).
// Rows contain payer and payment information as Name and Surname, bank number, payment amount, etc.
func ParsePayersFromFile(ss *excelize.File, sheet string) ([]model.Payer, error) {
	// Get all rows of the sheet as 2d array of strings
	rows, err := get(ss, sheet)
	if err != nil {
		return nil, err
	}

	// parsed rows stored here
	var payers []model.Payer

	for _, row := range rows[PayersRowStart-1:] {
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

		// format the cell row[7] that is an amount (column H)
		amount := normalizeAmount(row[7])

		// a row with the payer data
		payer := model.Payer{
			PersAcc:  strings.TrimSpace(row[0]),
			CHILDFIO: strings.TrimSpace(row[1]),
			Purpose:  strings.TrimSpace(row[2]),
			CBC:      strings.TrimSpace(row[3]),
			OKTMO:    strings.TrimSpace(row[4]),
			Sum:      amount,
		}

		payers = append(payers, payer)
	}
	return payers, nil
}

// ParsePayers opens Excel file and checks its validity calling OpenAndCheckSheets.
// If no error, then calls ParsePayersFromFile that performs parsing logic.
func ParsePayers(filepath string) ([]model.Payer, error) {
	sheet := PayersSheet // this is current settings sheet that must be in the spreadsheet
	ss, err := OpenAndCheckSheets(filepath, sheet)

	// Check if the file is valid and open it
	if err != nil {
		return nil, err
	}
	defer ss.Close()

	return ParsePayersFromFile(ss, sheet)
}

// ParseEmailFromFile parses all emails and payers full names from sheet.
// Each row contains two cells - Full Name and Email.
func ParseEmailFromFile(ss *excelize.File, sheet string) (map[string]string, error) {
	// Get all rows of the sheet as 2d array of strings
	rows, err := get(ss, sheet)
	if err != nil {
		return nil, err
	}

	// emails map will store emails in the format `full_name: email`
	emails := make(map[string]string)

	// missing will accumulate numbers of incomplete rows
	var missing []int

	for rowIdx, row := range rows[EmailsRowStart-1:] {
		// skip empty rows completely
		if len(row) == 0 || (strings.TrimSpace(row[0]) == "" && strings.TrimSpace(row[1]) == "") {
			continue
		}

		var fio, email string
		if len(row) > 0 {
			fio = strings.TrimSpace(row[0])
		}
		if len(row) > 1 {
			email = strings.TrimSpace(row[1])
		}

		// Check for missing FIO when email exists or vice versa
		if fio == "" && email != "" || fio != "" && email == "" {
			missing = append(missing, EmailsRowStart+rowIdx)
			continue
		}

		// if both fields are present → add to map
		emails[fio] = email
	}

	// if there were missing required fields, return them as error
	if len(missing) > 0 {
		return nil, &MissingEmailsError{MissingLines: missing}
	}

	return emails, nil
}

// ParseEmail opens Excel file and checks its validity calling OpenAndCheckSheets.
// If no error, then calls ParseEmailsFromFile that performs parsing logic.
func ParseEmail(filepath string) (map[string]string, error) {
	sheet := EmailsSheet // this is current settings sheet that must be in the spreadsheet
	ss, err := OpenAndCheckSheets(filepath, sheet)

	// Check if the file is valid and open it
	if err != nil {
		return nil, err
	}
	defer ss.Close()

	return ParseEmailFromFile(ss, sheet)
}

// normalizeAmount Appends "0" to ensure two decimal places, if there are only one decimal after the dot
// Example: 1200.5 -> 1200.50; 490.99 -> 490.99
func normalizeAmount(amount string) string {
	amount = strings.Replace(amount, ",", ".", 1)

	if strings.Contains(amount, ".") { // check if the cell is the decimal number
		splitted := strings.Split(amount, ".")

		tail := splitted[len(splitted)-1]

		if len(tail) == 1 { // if there are only one decimal after the dot
			amount = strings.TrimSpace(amount) + "0" // Append "0" to ensure two decimal places
		}
	} else {
		amount = strings.TrimSpace(amount) + ".00"
	}

	return amount
}
