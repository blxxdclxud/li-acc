package converter

var XLSXToPDF = Conversion{"xlsx", "pdf"}

func ExcelToPdf(inFilepath, outFilepath, publicKey, privateKey string) error {
	conv := NewConverter(ApiUrl, publicKey)

	err := conv.Convert(inFilepath, outFilepath, XLSXToPDF)
	return err
}
