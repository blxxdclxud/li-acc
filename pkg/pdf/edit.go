package pdf

import (
	"fmt"
	"li-acc/internal/errs"
	"li-acc/pkg/model"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/pdft"
)

const (
	DefaultFontName = "arial"
	DefaultFontPath = "static/fonts/Arial.ttf"

	FontStyleRegular = ""

	PaymentAmountFontSize    = 8
	PayerCredentialsFontSize = 9
)

// MultilineTextLineSpacing is the space between lines inside the FramePayerCredentialsTop and FramePayerCredentialsBottom frames.
// It is used since these credentials must be separated on 2 lines, but the pdft package does not split the text that does not fit the frame.
const MultilineTextLineSpacing = 12

// Canvas object if a wrapper for pdft.PDFt, that refers to one pdf object and provides following functionality:
// insert payer credentials into pdf receipt, insert payment amount and payment QR Code.
// DebugMode true, if it is needed to show frames Frame on the PDF receipt
type Canvas struct {
	pdf       *pdft.PDFt
	debugMode bool
}

// NewCanvasFromTemplate is a constructor for Canvas, loading given template.
func NewCanvasFromTemplate(pdfSrc, fontPath string, debugMode bool) (*Canvas, error) {
	if fontPath == "" {
		fontPath = DefaultFontPath
	}
	var pdf pdft.PDFt
	if err := pdf.Open(pdfSrc); err != nil {
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}

	// Upload specified font
	absFontPath, err := filepath.Abs(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve font path: %w", err)
	}
	if err := pdf.AddFont(DefaultFontName, absFontPath); err != nil {
		return nil, fmt.Errorf("failed to upload font: %w", err)
	}
	return &Canvas{pdf: &pdf, debugMode: debugMode}, nil
}

// GeneratePersonalReceipt fills the receipt pattern [pdfSrc] wilt payer's credentials [payerData] and
// payment amount in readable format [payerData.Sum].
// [fontPath] is the path to the preferred text font. Empty string causes using default font.
// Also inserts QR Code image redirecting to the fast payment by the same credentials.
func GeneratePersonalReceipt(pdfSrc, pdfDst, qrImgPath, fontPath string, payerData model.Payer, debugMode bool) error {
	// Initialize Canvas object
	canvas, err := NewCanvasFromTemplate(pdfSrc, fontPath, debugMode)
	if err != nil {
		return errs.Wrap(errs.System, "failed to create new PDF canvas from template file", err)
	}

	// read bytes of QR Code image
	img, err := os.ReadFile(qrImgPath)
	if err != nil {
		return errs.Wrap(errs.System, "failed to read a QR Code image file", err)
	}

	// fill canvas
	if err := canvas.Fill(payerData, img); err != nil {
		return errs.Wrap(errs.System, "failed to fill receipt", err)
	}

	// Save the filled file
	err = canvas.pdf.Save(pdfDst)
	if err != nil {
		return errs.Wrap(errs.System, fmt.Sprintf("failed to save filled pdf receipt `%s`", pdfDst), err)
	}

	return nil
}

func (c *Canvas) Fill(payer model.Payer, qrImg []byte) error {
	// Write payer's credentials to the receipt
	err := c.insertCredentials(payer)
	if err != nil {
		return fmt.Errorf("failed to insert credentials: %w", err)
	}
	// Write amount of the payment
	err = c.insertPaymentAmount(payer.Sum)
	if err != nil {
		return fmt.Errorf("failed to insert payment amount: %w", err)
	}
	// Draw QR Code to the receipt
	err = c.insertQrCode(qrImg)
	if err != nil {
		return fmt.Errorf("failed to insert qr code: %w", err)
	}
	return nil
}

func (c *Canvas) Save(path string) error {
	return c.pdf.Save(path)
}

// insertQrCode inserts a QR Code image of the payment, containing all the credentials provided in the receipt.
// Locates the image inside specified frame on the page.
func (c *Canvas) insertQrCode(qrImg []byte) error {
	x, y, w, h := FrameQrCode.InnerRect()
	return c.pdf.InsertImg(qrImg, 1, x, y, w, h)
}

// insertCredentials inserts a text, containing payer's credentials into a receipt.
// Sets the specified font size before and locates the text at specified position on the page.
func (c *Canvas) insertCredentials(payerData model.Payer) error {
	payerInfoText := prettifyCredentialsString(formatPayerInfo(payerData), "Назначение")

	// print the text inside specified frames on the page
	err := c.printToFrames([]Frame{FramePayerCredentialsTop, FramePayerCredentialsBottom}, payerInfoText, PayerCredentialsFontSize, true)
	if err != nil {
		return fmt.Errorf("failed to print payer credentials text: %w", err)
	}
	return nil
}

// insertPaymentAmount inserts a text, containing payment amount into a receipt.
// Sets the specified font size before and locates the text at specified position on the page.
func (c *Canvas) insertPaymentAmount(amount string) error {
	amountText := formatAmount(amount)

	// print the text inside specified frames on the page
	err := c.printToFrames([]Frame{FramePaymentAmountTop, FramePaymentAmountBottom}, amountText, PaymentAmountFontSize, false)
	if err != nil {
		return fmt.Errorf("failed to print amount text: %w", err)
	}
	return nil
}

func (c *Canvas) printToFrames(frames []Frame, text string, fontSize int, multiline bool) error {
	if err := c.pdf.SetFont(DefaultFontName, FontStyleRegular, fontSize); err != nil {
		return fmt.Errorf("failed to set font: %w", err)
	}

	for _, frame := range frames {
		if err := c.printText(frame, text, multiline); err != nil {
			return err
		}
	}
	return nil
}

// printText is a utility function that prints text inside a given frame on the PDF receipt page.
// It uses the `pdf` object to insert the provided `text` into the specified `frame`, horizontally centered.
// If `multiline` is true, the text is split by line breaks and printed line by line,
// increasing the y-coordinate by `MultilineTextLineSpacing` for each subsequent line.
// If `debugMode` is true, the frame’s borders are drawn on the PDF to visualize positioning.
func (c *Canvas) printText(frame Frame, text string, multiline bool) error {
	if c.debugMode {
		_ = frame.Debug(c.pdf, 1)
	}

	x, y, w, h := frame.InnerRect()

	if multiline {
		for _, line := range strings.Split(text, "\n") {
			if err := c.pdf.Insert(line, 1, x, y, w, h, pdft.Center, nil); err != nil {
				return err // сразу вернуть, а не продолжать
			}
			y += MultilineTextLineSpacing
		}
	} else {
		if err := c.pdf.Insert(text, 1, x, y, w, h, pdft.Center, nil); err != nil {
			return err
		}
	}

	return nil
}
