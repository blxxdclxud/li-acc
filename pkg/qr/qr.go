package qr

import (
	"github.com/skip2/go-qrcode"
	"strings"
)

const QRCodePath = "./tmp/qr-code.png"

func MakeQRData(qrData string, data []string) string {
	splitPattern := strings.Split(qrData, "~")
	for i := range splitPattern[:len(splitPattern)-1] {
		if i == 5 {
			if !strings.Contains(data[7], ".") && !strings.Contains(data[7], ",") {
				splitPattern[i] += strings.TrimSpace(data[7]) + "00"
			} else {
				trimmed := strings.TrimSpace(data[7])
				splitPattern[i] += strings.ReplaceAll(
					strings.ReplaceAll(trimmed, ".", ""),
					",",
					"",
				)
			}
		} else {
			splitPattern[i] += strings.ToUpper(strings.TrimSpace(data[i]))
		}
	}

	return strings.Join(splitPattern, "")
}

// GenerateQRCode generates new qr code from provided data qrData. It stores the image int the given path QRCodePath
func GenerateQRCode(qrData string) error {
	qr, err := qrcode.New(qrData, qrcode.Low)
	if err != nil {
		return err
	}

	qr.DisableBorder = true

	err = qr.WriteFile(80, // size of the square image of qr code
		QRCodePath)
	if err != nil {
		return err
	}

	return nil
}
