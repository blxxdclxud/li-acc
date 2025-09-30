package pdf

import (
	"fmt"
	"li-acc/pkg/model"
	"slices"
	"strings"
)

// lastIndex finds the last occurrence of an element in a slice
func lastIndex(slice []rune, value rune) int {
	// Reverse a copy of the slice
	reversed := slices.Clone(slice)
	slices.Reverse(reversed)

	// Find first occurrence in the reversed slice
	revIndex := slices.Index(reversed, value)
	if revIndex == -1 {
		return -1
	}

	// Convert reversed index to original index
	return len(slice) - 1 - revIndex
}

// prettifyCredentialsString makes credentials string fit the cell dedicated for it.
// since credentials string has the following format:
//
//	cred1: <>; cred2: <>; ...
//
// this function
//   - makes more spaces between credentials (after ';')
//   - splits string on two lines before `splitWord`, so the second line starts with it.
func prettifyCredentialsString(text, splitWord string) string {
	// split on two lines before splitWord
	text = strings.Replace(text, splitWord, "\n"+splitWord, 1)

	// number of spaces to separate credentials with
	extraSpaces := 2

	// make more space between credentials (each separated with ';') to make string more readable
	text = strings.Replace(
		text,
		";",
		";"+strings.Repeat(" ", extraSpaces),
		-1) // replace all occurrences

	return trimEachLine(text)
}

func trimEachLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

// formatAmount converts a monetary amount string into the format needed for PDF receipts.
// The amount can use either comma or dot as a decimal separator.
// Integers without decimals are treated as zero cents.
// Returns a string in the format "{rubles} руб. {kopeks} коп." with exactly two digits for kopeks.
// Examples:
//
//	"2000"    -> "2000 руб. 00 коп."
//	"3200,80" -> "3200 руб. 80 коп."
//	"100.05"  -> "100 руб. 05 коп."
func formatAmount(amount string) string {
	// Replace comma with dot for uniformity
	parts := strings.Split(strings.ReplaceAll(amount, ",", "."), ".")

	// rubles part
	rubles := parts[0]

	// kopeks part
	var kopeks string
	if len(parts) > 1 {
		kopeks = parts[1]
		if len(kopeks) == 1 {
			kopeks = kopeks + "0" // pad single digit with zero
		} else if len(kopeks) > 2 {
			kopeks = kopeks[:2] // truncate to two digits
		}
	} else {
		kopeks = "00"
	}

	return fmt.Sprintf("Сумма: %s руб. %s коп.", rubles, kopeks)
}

func formatPayerInfo(payerData model.Payer) string {
	return fmt.Sprintf("ЛС: %s; ФИО обучающегося: %s; Назначение: %s; КБК: %s; ОКТМО: %s",
		payerData.PersAcc, payerData.CHILDFIO, payerData.Purpose, payerData.CBC, payerData.OKTMO)
}
