package pdf

import (
	"fmt"
	"math"
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

func trimSpaceRunes(slice []rune) []rune {
	return []rune(strings.TrimSpace(string(slice)))
}

// alignText aligns text at center, given the length (width) in symbols of the text gap.
// Needed to show text in the receipt field properly.
// For example text "This is an example; just ignore", given length is 24, will be shown as
/*
	This is an example; just
             ignore
*/
func alignText(text string, length int) string {
	var result strings.Builder
	remaining := []rune(strings.TrimSpace(text))

	ratio := math.Ceil(float64(len(remaining)) / float64(length)) // the number of lines the text need to be divided to

	for i := 0; i < int(ratio); i++ {
		remaining = trimSpaceRunes(remaining)
		curr := []rune("") // text part on the i-th line

		// Find the nearest space at or after the specified length
		if len(remaining) < length {
			curr = remaining
		} else if spaceIdx := slices.Index(remaining[length:], ' '); spaceIdx != -1 { // if space found
			curr = remaining[:length+spaceIdx]
		} else {
			// No space found; take the rest of the string
			curr = remaining
		}

		// If the line is too long, truncate at the last space
		if len(curr) > length {
			lastSpaceIdx := lastIndex(curr, ' ')
			curr = curr[:lastSpaceIdx]
			remaining = trimSpaceRunes(remaining[lastSpaceIdx:])
		} else {
			// Remove the portion of the string already processed
			if len(remaining) > len(curr) {
				remaining = trimSpaceRunes(remaining[len(curr):])
			} else {
				remaining = []rune("")
			}
		}

		// Center the current line and append it to the result
		centered := fmt.Sprintf("%-*s", (length+83)/2, fmt.Sprintf("%-*s", length, string(curr)))
		result.WriteString(centered + "\n")
	}

	return result.String()
}

// formatAmount formats the payment amount to the needed pattern for receipt.
// Since the amount 2000,00 (or just 00) rubles should be represented as 200000 (the cents value just appended without delimiter).
// E.g. 3200,80 turns to 320080.
func formatAmount(amount string) string {
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
