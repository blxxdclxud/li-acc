//go:build integration

package integration

import (
	"li-acc/pkg/model"
	"li-acc/pkg/xls"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSettings_Integration(t *testing.T) {

	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join("testdata", "settings_valid.xlsm")

		org, err := xls.ParseSettings(path)
		require.NoError(t, err)

		require.Equal(t, "Школа АБВ (12377№7_) ТЕСТ ТЕСТ", org.Name)
		require.Equal(t, "7777777777777777777777", org.PersonalAcc)
		require.Equal(t, "ОТДЕЛЕНИЕ-НБ РЕСПУБЛИКА ТАТАРСТАН БАНКА ТЕСТ", org.BankName)
		require.Equal(t, "0987098980", org.BIC)
		require.Equal(t, "7777777777777777777777", org.CorrespAcc)
		require.Equal(t, "9959992999", org.PayeeINN)
		require.Equal(t, "3333333333", org.KPP)
		require.Equal(t, "CATEGORY=4", org.ExtraParams)
	})

	t.Run("invalid file (no settings page)", func(t *testing.T) {
		path := filepath.Join("testdata", "settings_invalid_no_page.xlsm")

		org, err := xls.ParseSettings(path)
		require.Error(t, err)
		require.Nil(t, org)

		require.Contains(t, err.Error(), "not contain sheet: Настройки")
	})

	t.Run("invalid file (missed parameters)", func(t *testing.T) {
		path := filepath.Join("testdata", "settings_invalid_missed_params.xlsm")

		errorContains := []string{"missing required parameters", "Наименование организации", "Расчетный счет"}

		org, err := xls.ParseSettings(path)
		require.Error(t, err)
		require.Nil(t, org)
		for _, substr := range errorContains {
			require.Contains(t, err.Error(), substr)
		}

	})
}

func TestParsePayers_Integration(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		path := filepath.Join("testdata", "payers_valid.xlsm")
		payers, err := xls.ParsePayers(path)
		require.NoError(t, err)
		require.Len(t, payers, 2) // между строками двух плательщиков есть пустые, их должен скипнуть

		require.Equal(t, model.Payer{
			PersAcc:  "123",
			CHILDFIO: "Иванов Иван",
			Purpose:  "Оплата",
			CBC:      "111",
			OKTMO:    "22222222",
			Sum:      "100.50", // уже готовая сумма
		}, payers[0])

		require.Equal(t, model.Payer{
			PersAcc:  "456",
			CHILDFIO: "Петров Петр",
			Purpose:  "Оплата",
			CBC:      "333",
			OKTMO:    "444",
			Sum:      "200.00", // в таблице просто 200
		}, payers[1])
	})

	t.Run("no rows after start", func(t *testing.T) {
		path := filepath.Join("testdata", "payers_empty.xlsm")
		payers, err := xls.ParsePayers(path)
		require.NoError(t, err)
		require.Empty(t, payers)
	})
}

func TestParseEmail_Integration(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		expectEmails  map[string]string
		expectErrPart []string
	}{
		{
			name:     "valid emails file",
			filename: "emails_valid.xlsx",
			expectEmails: map[string]string{
				"Иванов Иван": "ivanov@example.com",
				"Петров Петр": "petrov@example.com",
			},
		},
		{
			name:     "invalid emails file - missing fields",
			filename: "emails_invalid.xlsx",
			expectErrPart: []string{
				"missing email for FIO Сидоров Сидор",
				"missing FIO for email no-fio@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.filename)

			emails, err := xls.ParseEmail(path)

			if len(tt.expectErrPart) > 0 {
				require.Error(t, err)
				for _, part := range tt.expectErrPart {
					require.Contains(t, err.Error(), part)
				}
				require.Nil(t, emails)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectEmails, emails)
			}
		})
	}
}
