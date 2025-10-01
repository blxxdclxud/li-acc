package converter

import (
	"bytes"
	"encoding/json"
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

// ---- Tests for generateToken ----
func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name        string
		doFunc      func(*http.Request) (*http.Response, error)
		wantToken   string
		expectError bool
	}{
		{
			name: "success",
			doFunc: func(req *http.Request) (*http.Response, error) {
				bodyStruct := TokenResponse{Data: struct {
					AccessToken string `json:"accessToken"`
				}{AccessToken: "correctToken-123456-test"}}
				bodyBytes, _ := json.Marshal(bodyStruct)
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
				}, nil
			},
			wantToken:   "correctToken-123456-test",
			expectError: false,
		},
		{
			name: "bad response",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString("internal error")),
				}, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{client: &FakeClient{DoFunc: tt.doFunc}}
			token, err := c.generateToken("foo", "bar")

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, token)
			}
		})
	}
}

// ---- Tests for createTask ----
func TestCreateTask(t *testing.T) {
	tests := []struct {
		name        string
		doFunc      func(*http.Request) (*http.Response, error)
		wantTaskID  string
		expectError bool
	}{
		{
			name: "success",
			doFunc: func(req *http.Request) (*http.Response, error) {
				bodyStruct := CreateTaskResponse{Data: struct {
					TaskId string `json:"taskId"`
				}{TaskId: "task-123"}}
				bodyBytes, _ := json.Marshal(bodyStruct)
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
				}, nil
			},
			wantTaskID:  "task-123",
			expectError: false,
		},
		{
			name: "bad response",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString("internal error")),
				}, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{client: &FakeClient{DoFunc: tt.doFunc}, token: "dummy-token"}
			taskID, err := c.createTask(Conversion{"docx", "pdf"})

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantTaskID, taskID)
			}
		})
	}
}

// ---- Tests for uploadFile ----
func TestUploadFile(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "upload-test-*.txt")
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name        string
		doFunc      func(*http.Request) (*http.Response, error)
		wantFileKey string
		expectError bool
	}{
		{
			name: "success",
			doFunc: func(req *http.Request) (*http.Response, error) {
				bodyStruct := UploadFileResponse{Data: struct {
					FileKey string `json:"fileKey"`
				}{FileKey: "file-456"}}
				bodyBytes, _ := json.Marshal(bodyStruct)
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
				}, nil
			},
			wantFileKey: "file-456",
			expectError: false,
		},
		{
			name: "bad response",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString("upload failed")),
				}, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{client: &FakeClient{DoFunc: tt.doFunc}, token: "dummy-token"}
			fileKey, err := c.uploadFile("task-123", tmpFile.Name())

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantFileKey, fileKey)
			}
		})
	}
}

// ---- Tests for executeConversion ----
func TestExecuteConversion(t *testing.T) {
	tests := []struct {
		name        string
		doFunc      func(*http.Request) (*http.Response, error)
		expectError bool
	}{
		{
			name: "success",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString("{}")),
				}, nil
			},
			expectError: false,
		},
		{
			name: "bad response",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString("conversion error")),
				}, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{client: &FakeClient{DoFunc: tt.doFunc}, token: "dummy-token"}
			err := c.executeConversion("task-123")

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ---- Tests for getConvertedFileUrl ----
func TestGetConvertedFileUrl(t *testing.T) {
	tests := []struct {
		name        string
		doFunc      func(*http.Request) (*http.Response, error)
		wantURL     string
		expectError bool
	}{
		{
			name: "success",
			doFunc: func(req *http.Request) (*http.Response, error) {
				bodyStruct := GetConvertedResponse{Data: struct {
					DownloadUrl string `json:"downloadUrl"`
				}{DownloadUrl: "https://example.com/file.pdf"}}
				bodyBytes, _ := json.Marshal(bodyStruct)
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
				}, nil
			},
			wantURL:     "https://example.com/file.pdf",
			expectError: false,
		},
		{
			name: "bad response",
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString("failed")),
				}, nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Converter{client: &FakeClient{DoFunc: tt.doFunc}, token: "dummy-token"}
			url, err := c.getConvertedFileUrl("file-456")

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantURL, url)
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
