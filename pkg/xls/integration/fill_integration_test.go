//go:build integration

package integration

import (
	"fmt"
	"li-acc/pkg/model"
	"li-acc/pkg/xls"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestFillOrganizationParamsInReceipt(t *testing.T) {
	org := model.Organization{
		Name:        "Муниципальное автономное общеобразовательное учреждение \"Тест-тест №7\" Тест-Тестового района г.Казани (л/с 7777777777777777)",
		PersonalAcc: "20202020202020202020",
		BankName:    "ОТДЕЛЕНИЕ РЕСПУБЛИКА ТАТАРСТАН БАНКА ТЕСТА//ТТТ по Республике Татарстан г Казань",
		BIC:         "999999999",
		CorrespAcc:  "1111111111",
		PayeeINN:    "2222222222",
		KPP:         "888888888",
		ExtraParams: "test_extra_params",
	}

	tmpDir := t.TempDir()
	template := filepath.Join(tmpDir, "template.xlsx")

	// создаём минимальный шаблон
	f := excelize.NewFile()
	f.NewSheet(xls.BlankReceiptPatternSheet)
	if err := f.SaveAs(template); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	dstPath, err := xls.FillOrganizationParamsInReceipt(template, tmpDir, org)
	if err != nil {
		t.Fatalf("FillOrganizationParamsInReceipt returned error: %v", err)
	}

	// открываем результат
	res, err := excelize.OpenFile(dstPath)
	if err != nil {
		t.Fatalf("failed to open result: %v", err)
	}
	defer res.Close()

	tests := []struct {
		cell string
		want string
	}{
		{"C2", org.Name},
		{"C15", org.Name},
		{"C4", fmt.Sprintf("  ИНН %s КПП %s%s%s", org.PayeeINN, org.KPP, strings.Repeat(" ", 25), org.PersonalAcc)},
		{"C17", fmt.Sprintf("  ИНН %s КПП %s%s%s", org.PayeeINN, org.KPP, strings.Repeat(" ", 25), org.PersonalAcc)},
		{"C6", fmt.Sprintf("БИК %s (%s)", org.BIC, org.BankName)},
		{"C19", fmt.Sprintf("БИК %s (%s)", org.BIC, org.BankName)},
	}

	for _, tt := range tests {
		got, _ := res.GetCellValue(xls.BlankReceiptPatternSheet, tt.cell)
		if got != tt.want {
			t.Errorf("%s: want %q, got %q", tt.cell, tt.want, got)
		}
	}
}
