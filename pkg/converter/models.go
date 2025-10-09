package converter

type ProcessConversionResponse struct {
	Data struct {
		FileInfo []struct {
			DownloadUrl string `json:"downloadUrl"`
			Status      string `json:"status"`
		} `json:"fileInfoDTOList"`
	} `json:"data"`
}
