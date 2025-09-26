package xls

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"li-acc/pkg/model"
)

func TestParseSettingsFromFile(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]string
		wantOrg       *model.Organization
		wantError     bool
		errorContains []string
	}{
		{
			name: "valid data",
			data: map[string]string{
				"Наименование организации": "ООО Ромашка",
				"Расчетный счет":           "1234567890",
				"Наименование банка":       "Сбер",
				"БИК":                      "123456",
				"Корреспондентский счет":   "0987654321",
				"ИНН": "7701234567",
				"КПП": "770101001",
				"Дополнительные параметры ДШК": "extra",
				"Шаблон":     "pattern",
				"Код услуги": "code",
				"Каталог для выгрузки реестра": "catalog",
			},
			wantOrg: &model.Organization{
				Name:        "ООО Ромашка",
				PersonalAcc: "1234567890",
				BankName:    "Сбер",
				BIC:         "123456",
				CorrespAcc:  "0987654321",
				PayeeINN:    "7701234567",
				KPP:         "770101001",
				ExtraParams: "extra",
			},
			wantError: false,
		},
		{
			name: "missing required keys",
			data: map[string]string{
				// "Наименование организации" — removed
				// "Расчетный счет" — removed
				"Наименование банка": "Сбер",
				"БИК":                "123456",
				"Корреспондентский счет": "0987654321",
				"ИНН": "7701234567",
				"КПП": "770101001",
				"Дополнительные параметры ДШК": "extra",
			},
			wantOrg:       nil,
			wantError:     true,
			errorContains: []string{"missing required parameters", "Наименование организации", "Расчетный счет"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := excelize.NewFile()
			sheet := SettingsSheet
			index, _ := f.NewSheet(sheet)
			f.SetActiveSheet(index)

			// Заполняем A2:B12 данными
			row := SettingsRowStart
			for k, v := range tt.data {
				cellA, _ := excelize.CoordinatesToCellName(1, row)
				cellB, _ := excelize.CoordinatesToCellName(2, row)
				f.SetCellValue(sheet, cellA, k)
				f.SetCellValue(sheet, cellB, v)
				row++
			}

			org, err := ParseSettingsFromFile(f, sheet)
			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, org)
				for _, substr := range tt.errorContains {
					require.Contains(t, err.Error(), substr)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantOrg, org)
			}
		})
	}
}

func TestParsePayersFromFile(t *testing.T) {
	f := excelize.NewFile()
	sheet := PayersSheet
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)

	// Старт с 7-й строки
	f.SetCellValue(sheet, "A7", "12345")       // PersAcc
	f.SetCellValue(sheet, "B7", "Иванов И.И.") // CHILDFIO
	f.SetCellValue(sheet, "C7", "Назначение")  // Purpose
	f.SetCellValue(sheet, "D7", "123")         // CBC
	f.SetCellValue(sheet, "E7", "456")         // OKTMO
	f.SetCellValue(sheet, "H7", "100.5")       // Sum (col H = index 7)
	f.SetCellValue(sheet, "A8", "")            // пустая строка для проверки пропуска

	payers, err := ParsePayersFromFile(f, sheet)
	require.NoError(t, err)
	require.Len(t, payers, 1)
	require.Equal(t, model.Payer{
		PersAcc:  "12345",
		CHILDFIO: "Иванов И.И.",
		Purpose:  "Назначение",
		CBC:      "123",
		OKTMO:    "456",
		Sum:      "100.50", // проверка на нормализацию десятичной
	}, payers[0])
}

func TestParseEmailFromFile(t *testing.T) {
	f := excelize.NewFile()
	sheet := EmailsSheet
	index, _ := f.NewSheet(sheet)
	f.SetActiveSheet(index)

	// начинаем с 2-й строки
	f.SetCellValue(sheet, "A2", "Иванов И.И.")
	f.SetCellValue(sheet, "B2", "ivanov@example.com")

	emails, err := ParseEmailFromFile(f, sheet)
	require.NoError(t, err)
	require.Len(t, emails, 1)
	require.Equal(t, "ivanov@example.com", emails["Иванов И.И."])
}
