package converter

import "li-acc/config"

var XLSXToPDF = Conversion{"xlsx", "pdf"}

func ExcelToPdf(inFilepath, outFilepath string) error {
	conv, err := NewConverter(apiUrl,
		config.LoadConfig().ConvertAPI.PublicKey,
		config.LoadConfig().ConvertAPI.PrivateKey)
	if err != nil {
		return err
	}

	err = conv.Convert(inFilepath, outFilepath, XLSXToPDF)
	return err
}
