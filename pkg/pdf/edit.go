package pdf

import (
	"fmt"
	"github.com/signintech/pdft"
	"li-acc/pkg/model"
	"os"
	"strings"
)

// Canvas object if a wrapper for pdft.PDFt, that refers to one pdf object and provides following functionality:
// insert payer credentials into pdf receipt, insert payment amount and payment QR Code.
type Canvas struct {
	pdf *pdft.PDFt
}

// insertCredentials inserts a text, containing payer's credentials into a receipt.
// Sets the specified font size before and locates the text at specified position on the page.
func (c *Canvas) insertCredentials(payerData model.Payer) {
	_ = c.pdf.SetFont("arial", "", 9)

	payerInfoText := alignText(
		fmt.Sprintf("ЛС: %s; ФИО обучающегося: %s; Назначение: %s; КБК: %s; ОКТМО: %s",
			payerData.PersAcc, payerData.CHILDFIO, payerData.Purpose, payerData.CBC, payerData.OKTMO),
		80)

	// print the text on specified position on the page
	printPayerCredentials(c.pdf, payerInfoText,
		115.0) // y-coordinate (from the top of the page) of the first block of the payer credentials
	printPayerCredentials(c.pdf, payerInfoText,
		306.0) // y-coordinate (from the top of the page) of the second block of the payer credentials
}

// insertPaymentAmount inserts a text, containing payment amount into a receipt.
// Sets the specified font size before and locates the text at specified position on the page.
func (c *Canvas) insertPaymentAmount(amount string) {
	_ = c.pdf.SetFont("arial", "", 8)

	// print the text on specified position on the page
	printPaymentAmount(c.pdf, amount,
		151.0) // y-coordinate (from the top of the page) of the first block of the payment amount
	printPaymentAmount(c.pdf, amount,
		343.0) // y-coordinate (from the top of the page) of the second block of the payment amount
}

// insertCredentials inserts a QR Code image of the payment, containing all the credentials provided in the receipt.
// Locates the image at specified position on the page.
func (c *Canvas) insertQrCode(qrImg []byte) {
	// Qr Code image size is 120x120
	_ = c.pdf.InsertImg(qrImg, 1, 37, 270, 120, 120)
}

// GeneratePersonalReceipt fills the receipt pattern [pdfSrc] wilt payer's credentials [payerData] and
// payment amount in readable format [payerData.Sum].
// Also inserts QR Code image redirecting to the fast payment by the same credentials.
func GeneratePersonalReceipt(pdfSrc, pdfDst, qrImgPath string, payerData model.Payer) error {
	// Initialize pdf object to perform operations with opened PDF
	var pdf pdft.PDFt

	// Open pattern PDF file with already filled Organization information
	err := pdf.Open(pdfSrc)
	if err != nil {
		return fmt.Errorf("failed to open PDF receipt pattern `%s`: %w", pdfSrc, err)
	}

	// Initialize Canvas object
	c := Canvas{pdf: &pdf}

	// Upload specified font
	err = c.pdf.AddFont("arial", "./static/fonts/Arial.ttf")
	if err != nil {
		return fmt.Errorf("failed to upload a font for pdf: %w", err)
	}

	// read bytes of QR Code image
	img, err := os.ReadFile(qrImgPath)
	if err != nil {
		return fmt.Errorf("failed to read a QR Code image file: %w", err)
	}

	// Write payer's credentials to the receipt
	c.insertCredentials(payerData)
	// Draw QR Code to the receipt
	c.insertQrCode(img)

	// Save the filled file
	err = pdf.Save(pdfDst)
	if err != nil {
		return fmt.Errorf("failed to save filled pdf receipt `%s`: %w", pdfDst, err)
	}

	return err
}

// printPayerCredentials is util function, responsible for printing payer's credentials on the receipt page.
// Uses `pdf` object to add text `text` to the y-position `y`, aligned to the left.
// Splits the text to separate lines and prints them separately, increasing y-coordinate for each next line.
func printPayerCredentials(pdf *pdft.PDFt, text string, y float64) {
	for _, line := range strings.Split(text, "\n") {
		_ = pdf.Insert(line, 1, 166, y, 150, 50, pdft.Left, nil)
		y += 12
	}
}

// printPayerCredentials is util function, responsible for printing payment amount in human-readable format on the receipt page.
// Uses `pdf` object to add `amount` to the y-position `y`, aligned to the left.
func printPaymentAmount(pdf *pdft.PDFt, amount string, y float64) {
	_ = pdf.Insert(
		fmt.Sprintf("Сумма: %s", formatAmount(amount)),
		1, 305, y, 150, 50, pdft.Left, nil,
	)
}
