package qr

import (
	"fmt"
	"strings"
)

// FormatAmount formats the payment amount to the needed pattern for receipt.
// Since the amount 2000,00 (or just 00) rubles should be represented as 200000 (the cents value just appended without delimiter).
// E.g. 3200,80 turns to 320080.
func FormatAmount(amount string) string {
	// No matter if delimiter is ',' or '.', it means the same. Then split int part and decimal part
	fmtAmount := strings.Split(
		strings.ReplaceAll(amount, ",", "."),
		".",
	)
	if len(fmtAmount) == 1 { // if there is no decimal part
		fmtAmount = append(fmtAmount, "00") // make the cents value default (00 cents)
	}

	return fmt.Sprintf("%s руб. %s коп.", fmtAmount[0], fmtAmount[1])
}
