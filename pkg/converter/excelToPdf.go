package converter

var XLSXToPDF = Conversion{"xlsx", "pdf"}

func ExcelToPdf(inFilepath, outFilepath string) error {
	conv, err := NewConverter()
	if err != nil {
		return err
	}

	err = conv.Convert(inFilepath, outFilepath, XLSXToPDF)
	return err
}
