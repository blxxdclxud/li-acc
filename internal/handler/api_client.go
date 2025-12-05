// internal/handler/api_client.go
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"li-acc/internal/model"
	"mime/multipart"
	"net/http"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    baseURL + ApiRoutesGroup,
		httpClient: &http.Client{},
	}
}

// Загрузка файла с плательщиками
func (c *APIClient) UploadPayers(filename string, fileData io.Reader) (*PayersFileUploadResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, fileData); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+ApiEndpointUploadPayers, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Читаем error response
		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("%d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", errResp["error"])
	}

	var result PayersFileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Загрузка файла с почтами
func (c *APIClient) UploadEmails(filename string, fileData io.Reader) (*EmailsFileUploadResponseSuccess, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, fileData); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+ApiEndpointUploadEmails, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("%d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s", errResp["error"])
	}

	var result EmailsFileUploadResponseSuccess
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Получение истории
func (c *APIClient) GetHistory() ([]model.File, error) {
	resp, err := c.httpClient.Get(c.baseURL + ApiEndpointGetHistory)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}

	var result FilesHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Files, nil
}
