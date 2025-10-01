//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"li-acc/pkg/converter"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var conversion = converter.Conversion{From: "docx", To: "pdf"}

// --- Фейковый сервер, который эмулирует весь пайплайн ---
func newFakeConvertServer(t *testing.T, downloadContent []byte) *httptest.Server {
	mux := http.NewServeMux()

	createTaskEndpoint := conversion.Endpoint()

	mux.HandleFunc(converter.EndpointToken, func(w http.ResponseWriter, r *http.Request) {
		resp := converter.TokenResponse{Data: struct {
			AccessToken string `json:"accessToken"`
		}{AccessToken: "token-123"}}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc(createTaskEndpoint, func(w http.ResponseWriter, r *http.Request) {
		resp := converter.CreateTaskResponse{Data: struct {
			TaskId string `json:"taskId"`
		}{TaskId: "task-123"}}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc(converter.EndpointUploadFile, func(w http.ResponseWriter, r *http.Request) {
		resp := converter.UploadFileResponse{Data: struct {
			FileKey string `json:"fileKey"`
		}{FileKey: "file-456"}}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc(converter.EndpointConvert, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	})

	ts := httptest.NewServer(mux)

	fileDownloadUrl := "/download/file-456"
	mux.HandleFunc(converter.EndpointGetConverted, func(w http.ResponseWriter, r *http.Request) {
		resp := converter.GetConvertedResponse{Data: struct {
			DownloadUrl string `json:"downloadUrl"`
		}{DownloadUrl: ts.URL + fileDownloadUrl}}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc(fileDownloadUrl, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(downloadContent) // вернём тестовый байтовый файл
	})

	return ts
}

func TestConvert_E2E(t *testing.T) {
	// Контент, который должен "скачаться"
	expectedContent := []byte("this is converted pdf data")

	// Поднимаем fake server
	server := newFakeConvertServer(t, expectedContent)
	defer server.Close()

	// Патчим URL'ы на наш тестовый сервер
	c, err := converter.NewConverter(server.URL, "public", "private")
	require.NoError(t, err)

	// Входной и выходной файлы
	inFile, _ := os.CreateTemp("", "in-test-*.txt")
	defer os.Remove(inFile.Name())
	outFile := inFile.Name() + ".out"

	// Вызываем сам метод
	err = c.Convert(inFile.Name(), outFile, conversion)
	require.NoError(t, err)

	// Проверяем, что на диске действительно то, что отдал сервер
	got, err := os.ReadFile(outFile)
	require.NoError(t, err)
	require.Equal(t, expectedContent, got)
}
