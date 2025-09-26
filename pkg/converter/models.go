package converter

type tokenResponse struct {
	Data struct {
		AccessToken string `json:"accessToken"`
	} `json:"data"`
}

type createTaskResponse struct {
	Data struct {
		TaskId string `json:"taskId"`
	} `json:"data"`
}

type uploadFileResponse struct {
	Data struct {
		FileKey string `json:"fileKey"`
	} `json:"data"`
}

type getConvertedResponse struct {
	Data struct {
		DownloadUrl string `json:"downloadUrl"`
	} `json:"data"`
}
