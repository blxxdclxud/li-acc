package xls

import (
	"fmt"
	"li-acc/internal/errs"
	"li-acc/pkg/model"
	"strings"

	"github.com/xuri/excelize/v2"
)

const BlankReceiptPatternSheet = "receipt" // The only sheet in the .xls file that stores receipt pattern

var FillerReceiptPatternFileName = "./tmp/receipt_pattern.xlsx"

// WriteToCells writes sender's information given as `orgData` into given excelize.File.
// It does not care about paths of the file, etc.
func WriteToCells(f *excelize.File, orgData model.Organization) error {
	set := func(cell, value string) error {
		if err := f.SetCellStr(BlankReceiptPatternSheet, cell, value); err != nil {
			return errs.Wrap(errs.System,
				fmt.Sprintf("failed to set cell %s on sheet %s", cell, BlankReceiptPatternSheet),
				err,
			)
		}
		return nil
	}

	// row1
	row1 := orgData.Name
	if err := set("C2", row1); err != nil {
		return err
	}
	if err := set("C15", row1); err != nil {
		return err
	}

	// row2
	row2 := fmt.Sprintf(
		"  ИНН %s КПП %s%s%s",
		orgData.PayeeINN, orgData.KPP, strings.Repeat(" ", 25), orgData.PersonalAcc,
	)
	if err := set("C4", row2); err != nil {
		return err
	}
	if err := set("C17", row2); err != nil {
		return err
	}

	// row3
	row3 := fmt.Sprintf("БИК %s (%s)", orgData.BIC, orgData.BankName)
	if err := set("C6", row3); err != nil {
		return err
	}
	if err := set("C19", row3); err != nil {
		return err
	}

	return nil
}

// FillOrganizationParamsInReceipt gets 'filepath' that is path to the .xls file that stores the blank pattern of the receipt.
// Then calls fillParams() that fills this receipt with sender's information.
// This information is stored in 'orgData'.
func FillOrganizationParamsInReceipt(filepath string, org model.Organization) error {
	f, err := excelize.OpenFile(filepath)
	if err != nil {
		return errs.WrapIOError("open excel file - pattern of the receipt", filepath, err)
	}
	defer f.Close()

	if err := WriteToCells(f, org); err != nil {
		return errs.WrapIOError("write organisation parameters into file", filepath, err)
	}

	if err := f.SaveAs(FillerReceiptPatternFileName); err != nil {
		return errs.WrapIOError("save excel file with organisation data", FillerReceiptPatternFileName, err)
	}
	return nil
}
