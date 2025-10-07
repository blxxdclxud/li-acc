package converter

var XLSXToPDF = Conversion{"xlsx", "pdf"}

func ExcelToPdf(inFilepath, outFilepath, publicKey, privateKey string) error {
	conv, err := NewConverter(ApiUrl,
		publicKey,
		privateKey)
	if err != nil {
		return err
	}

	err = conv.Convert(inFilepath, outFilepath, XLSXToPDF)
	return err
}
