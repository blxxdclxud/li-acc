package converter

type TokenResponse struct {
	Data struct {
		AccessToken string `json:"accessToken"`
	} `json:"data"`
}

type CreateTaskResponse struct {
	Data struct {
		TaskId string `json:"taskId"`
	} `json:"data"`
}

type UploadFileResponse struct {
	Data struct {
		FileKey string `json:"fileKey"`
	} `json:"data"`
}

type GetConvertedResponse struct {
	Data struct {
		DownloadUrl string `json:"downloadUrl"`
	} `json:"data"`
}
