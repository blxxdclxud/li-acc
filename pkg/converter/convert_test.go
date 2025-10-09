package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type FakeClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (f FakeClient) Do(req *http.Request) (*http.Response, error) {
	return f.DoFunc(req)
}

// ---- Tests for processConversion ----
func TestProcessConversion(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "conversion-test-*.txt")
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name        string
		doFunc      func(*http.Request) (*http.Response, error)
		expectError bool
		wantResp    ProcessConversionResponse
	}{
		{
			name: "success",
			doFunc: func(req *http.Request) (*http.Response, error) {
				require.Equal(t, "dummy-key", req.Header.Get("x-api-key"))
				require.Contains(t, req.Header.Get("Content-Type"), "multipart/form-data")

				bodyStruct := ProcessConversionResponse{
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
								DownloadUrl: "file-download/134253566",
								Status:      "success",
							},
						},
					},
				}
				bodyBytes, _ := json.Marshal(bodyStruct)

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
				}, nil
			},
			wantResp: ProcessConversionResponse{
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
							DownloadUrl: "file-download/134253566",
							Status:      "success",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "server error",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString("internal server error")),
				}, nil
			},
			expectError: true,
		},
		{
			name: "http client error",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("connection refused")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{
				client: &FakeClient{DoFunc: tt.doFunc},
				apiKey: "dummy-key",
			}

			conversion := Conversion{
				From: "docx", To: "pdf",
			}

			resp, err := c.processConversion(tmpFile.Name(), conversion)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantResp, resp)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		expectError bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("hello world"))
			},
			expectError: false,
		},
		{
			name: "bad status code",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal error", http.StatusInternalServerError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			tmpFile, _ := os.CreateTemp("", "download-test-*.txt")
			defer os.Remove(tmpFile.Name())

			err := downloadFile(server.URL, tmpFile.Name())

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				content, _ := os.ReadFile(tmpFile.Name())
				require.Equal(t, "hello world", string(content))
			}
		})
	}
}
