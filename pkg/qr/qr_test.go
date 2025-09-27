package qr

import (
	"strings"
	"testing"

	"li-acc/pkg/model"

	"github.com/stretchr/testify/require"
)

// ---- Fixtures ----

var orgFix = model.Organization{
	Name:        "Муниципальное автономное общеобразовательное учреждение \"Тест-тест №7\" Тест-Тестового района г.Казани (л/с 7777777777777777)",
	PersonalAcc: "20202020202020202020",
	BankName:    "ОТДЕЛЕНИЕ РЕСПУБЛИКА ТАТАРСТАН БАНКА ТЕСТА//ТТТ по Республике Татарстан г Казань",
	BIC:         "999999999",
	CorrespAcc:  "1111111111",
	PayeeINN:    "2222222222",
	KPP:         "888888888",
	ExtraParams: "test_extra_params",
}

var payerFix = model.Payer{
	PersAcc:  "123456",
	CHILDFIO: "TEST NAME Surname",
	Purpose:  "test",
	CBC:      "123454656543",
	OKTMO:    "9879098",
	Sum:      "10150.40",
}

// ---- Tests ----

// 1. Unit test: GetPayersQrDataString
func TestGetPayersQrDataString(t *testing.T) {
	qr := NewQrPattern(orgFix)
	qrStr := qr.GetPayersQrDataString(payerFix)

	// must contain payment header
	require.Contains(t, qrStr, "ST00012")

	// organization data
	require.Contains(t, qrStr, "Name="+orgFix.Name)
	require.Contains(t, qrStr, "PersonalAcc="+orgFix.PersonalAcc)
	require.Contains(t, qrStr, "|"+orgFix.ExtraParams)

	// payer data (uppercase transformation)
	require.Contains(t, qrStr, "CHILDFIO="+strings.ToUpper(payerFix.CHILDFIO))
	require.Contains(t, qrStr, "Purpose="+strings.ToUpper(payerFix.Purpose))
	require.Contains(t, qrStr, "Sum="+strings.Replace(payerFix.Sum, ".", "", 1)) // should be correctly formatted with removed point

	// required LASTNAME field always present
	require.Contains(t, qrStr, "LASTNAME=")

	// extra param should be last
	require.Contains(t, qrStr, orgFix.ExtraParams)
}

// 2. Unit test: structFieldsToString
func TestStructFieldsToString(t *testing.T) {
	org := model.Organization{
		Name:        "Test Org",
		PersonalAcc: "111",
		ExtraParams: "zzz",
	}
	payer := model.Payer{
		PersAcc:  "222",
		CHILDFIO: "Ivanov",
	}

	partsOrg := structFieldsToString(org)
	partsPayer := structFieldsToString(payer)

	// all fields represented
	require.Contains(t, partsOrg, "Name="+org.Name)
	require.Contains(t, partsOrg, "PersonalAcc="+org.PersonalAcc)
	// ExtraParams must be missed, since in model.Organization tag 'includeQr' is false
	require.NotContains(t, partsOrg, "ExtraParams")
	require.NotContains(t, partsOrg, org.ExtraParams)

	require.Contains(t, partsPayer, "PersAcc="+payer.PersAcc)
	require.Contains(t, partsPayer, "CHILDFIO="+payer.CHILDFIO)
}

// Edge case: empty payer
func TestGetPayersQrDataString_EmptyPayer(t *testing.T) {
	qr := NewQrPattern(orgFix)
	qrStr := qr.GetPayersQrDataString(model.Payer{})

	// should still contain org data
	require.Contains(t, qrStr, "Name="+orgFix.Name)
	// LASTNAME must be present even if payer empty
	require.Contains(t, qrStr, "LASTNAME=")
}
