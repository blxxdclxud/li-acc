package qr

import (
	"fmt"
	"image/jpeg"
	"li-acc/pkg/model"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/skip2/go-qrcode"
)

// QrCode represents the data that QrCode of a payment contains, such as Receiver (Organization) and Sender (Payer) contains.
type QrCode struct {
	model.Organization        // The part of QR code containing information about Organization
	model.Payer               // The part of QR code containing information about Payer
	LASTNAME           string // Needed in QR pattern, empty string
}

// NewQrPattern initializes new QrCode object with given model.Organization data
// It is a pattern, since all payment receiver (organization) information is set up.
// Each payer enters only his information, and payment QR will be fully ready.
func NewQrPattern(orgData model.Organization) *QrCode {
	return &QrCode{
		Organization: orgData,
		Payer:        model.Payer{},
	}
}

// GetPayersQrDataString inserts model.Payer's data into QrCode pattern, editing some fields in the beginning.
// Then it converts all fields of the QrData structure into the string "ST00012|Field1=Value1|Field2=Value2|...|Value_Of_ExtraParams"
func (q QrCode) GetPayersQrDataString(payerData model.Payer) string {
	// edit some fields
	payerData.CHILDFIO = strings.ToUpper(payerData.CHILDFIO)
	payerData.Purpose = strings.ToUpper(payerData.Purpose)
	payerData.Sum = formatAmount(payerData.Sum)

	// insert payer's data into qr data
	q.Payer = payerData

	var parts []string

	// Payment pattern
	parts = append(parts, "ST00012")

	// Payment parameters below in the format "Param=Value"
	parts = append(parts, structFieldsToString(q.Organization)...)
	parts = append(parts, structFieldsToString(q.Payer)...)
	parts = append(parts, fmt.Sprintf("%s=%v", "LASTNAME", ""))

	// Extra parameter, some meta information in format "Value", without parameter as above
	parts = append(parts, q.ExtraParams)

	return strings.Join(parts, "|")
}

// GenerateQRCode generates new qr code from provided data qrData. It stores the image at the given path outPath.
// If the outPath's base does not exist, it creates all needed directories to store file.
func (q QrCode) GenerateQRCode(qrData, outPath string) error {
	// create directory of the output file, if it does not exist
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("failed to open ")
	}

	qr, err := qrcode.New(qrData, qrcode.Low)
	if err != nil {
		return err
	}

	qr.DisableBorder = true

	img := qr.Image(80)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 100})

	if err != nil {
		return err
	}

	return nil
}

// structFieldsToString gets the struct of the type T (that is either model.Payer or model.Organization) and
// converts each filed to string of the format "Field=Value", except those, that have tag `includeQr:"false"`. Tag is true by default.
// Returns an array of string-formatted fields.
func structFieldsToString[T model.Payer | model.Organization](s T) []string {
	var parts []string

	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).Tag.Get("includeQr") == "false" {
			continue
		}

		fieldName := t.Field(i).Name
		fieldValue := v.Field(i)

		parts = append(parts, fmt.Sprintf("%s=%v", fieldName, fieldValue))
	}

	return parts
}
