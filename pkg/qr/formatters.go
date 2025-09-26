package qr

import "strings"

// formatAmount formats the payment amount value to fit the QrCode data format.
// Shortly, amount of the format 1230.00 must look as 123000.
func formatAmount(amount string) string {
	// No matter if delimiter is ',' or '.', it means the same. Then split int part and decimal part
	fmtAmount := strings.Split(
		strings.ReplaceAll(amount, ",", "."),
		".",
	)
	if len(fmtAmount) == 1 { // if there is no decimal part
		fmtAmount = append(fmtAmount, "00") // make the cents value default (00 cents)
	} else if len(fmtAmount[1]) == 1 { // if the amount in the format 230.4, make it 230.40
		fmtAmount[1] += "0"
	}

	return strings.Join(fmtAmount, "")
}
