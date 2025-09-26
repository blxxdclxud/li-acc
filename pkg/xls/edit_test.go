package xls

import (
	"fmt"
	"li-acc/pkg/model"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestWriteToCells(t *testing.T) {
	data := model.Organization{
		Name:        "Муниципальное автономное общеобразовательное учреждение \"Тест-тест №7\" Тест-Тестового района г.Казани (л/с 7777777777777777)",
		PersonalAcc: "20202020202020202020",
		BankName:    "ОТДЕЛЕНИЕ РЕСПУБЛИКА ТАТАРСТАН БАНКА ТЕСТА//ТТТ по Республике Татарстан г Казань",
		BIC:         "999999999",
		CorrespAcc:  "1111111111",
		PayeeINN:    "2222222222",
		KPP:         "888888888",
		ExtraParams: "test_extra_params",
	}

	f := excelize.NewFile()
	sheet := BlankReceiptPatternSheet
	f.NewSheet(sheet)

	if err := WriteToCells(f, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := map[string]string{
		"C2":  data.Name,
		"C15": data.Name,
		"C4":  fmt.Sprintf("  ИНН %s КПП %s%s%s", data.PayeeINN, data.KPP, strings.Repeat(" ", 25), data.PersonalAcc),
		"C17": fmt.Sprintf("  ИНН %s КПП %s%s%s", data.PayeeINN, data.KPP, strings.Repeat(" ", 25), data.PersonalAcc),
		"C6":  fmt.Sprintf("БИК %s (%s)", data.BIC, data.BankName),
		"C19": fmt.Sprintf("БИК %s (%s)", data.BIC, data.BankName),
	}

	for cell, want := range tests {
		got, err := f.GetCellValue(sheet, cell)
		if err != nil {
			t.Errorf("GetCellValue(%s): %v", cell, err)
			continue
		}
		if got != want {
			t.Errorf("cell %s: want %q, got %q", cell, want, got)
		}
	}
}
