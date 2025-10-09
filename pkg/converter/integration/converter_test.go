//go:build integration

package integration

import (
	"encoding/json"
	"li-acc/pkg/converter"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var conversion = converter.Conversion{From: "docx", To: "pdf"}

// --- Фейковый сервер, который эмулирует весь пайплайн ---
func newFakeConvertServer(t *testing.T, downloadContent []byte) *httptest.Server {
	mux := http.NewServeMux()

	ts := httptest.NewServer(mux)

	fileDownloadUrl := "/download/file-456"

	mux.HandleFunc(conversion.ConversionFormatEndpoint(), func(w http.ResponseWriter, r *http.Request) {
		resp := converter.ProcessConversionResponse{
			Data: struct {
				FileInfo []struct {
					DownloadUrl string `json:"downloadUrl"`
					Status      string `json:"status"`
				} `json:"fileInfoDTOList"`
			}{
				FileInfo: []struct {
					DownloadUrl string `json:"downloadUrl"`
					Status      string `json:"status"`
				}{
					{
						DownloadUrl: ts.URL + fileDownloadUrl,
						Status:      "success",
					},
				},
			},
		}
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
	c := converter.NewConverter(server.URL, "public")

	// Входной и выходной файлы
	inFile, _ := os.CreateTemp("", "in-test-*.txt")
	defer os.Remove(inFile.Name())
	outFile := inFile.Name() + ".out"

	// Вызываем сам метод
	err := c.Convert(inFile.Name(), outFile, conversion)
	require.NoError(t, err)

	// Проверяем, что на диске действительно то, что отдал сервер
	got, err := os.ReadFile(outFile)
	require.NoError(t, err)
	require.Equal(t, expectedContent, got)
}

func TestConvert_RealCase(t *testing.T) {
	xlsFilePath := "./testdata/receipt_pattern.xlsx"

	if err := godotenv.Load(".env.local"); err != nil {
		t.Skip("no .env.local found, skipping real data test")
	}

	pubKey, secKey := os.Getenv("PUBLIC"), os.Getenv("SECRET")
	if pubKey == "" || secKey == "" {
		t.Skip("environment variables are empty")
	}

	outDir := "./testdata/out"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Skipf("failed to create directory %s: %v", outDir, err)
	}

	dst := path.Join(outDir, "converted.pdf")

	err := converter.ExcelToPdf(xlsFilePath, dst, pubKey, secKey)
	require.NoError(t, err)

	_, err = os.Stat(dst)
	require.NoError(t, err)
}
