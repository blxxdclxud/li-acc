package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"li-acc/internal/handler"
	"li-acc/internal/mocks"
)

// helper to create multipart file upload request
func newMultipartRequest(t *testing.T, fieldName, filename, content string) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, err
	}

	if _, err := strings.NewReader(content).WriteTo(part); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/settings/upload-emails", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestUploadEmailsFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.SettingsService)
	// Expect ProcessEmailsFile called with any context, filename, and byte slice, returning nil error
	mockSvc.On("ProcessEmailsFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

	h := handler.NewSettingsHandler(mockSvc)

	// Prepare multipart request with dummy Excel file content
	req, err := newMultipartRequest(t, "file", "test.xlsx", "dummy Excel content")
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	h.UploadEmailsFile(c)

	// Assert success response
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "file processed successfully", resp["message"])

	mockSvc.AssertExpectations(t)
}

func TestUploadEmailsFile_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.SettingsService)
	// Simulate ProcessEmailsFile returns error
	mockSvc.On("ProcessEmailsFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(errors.New("processing error"))

	h := handler.NewSettingsHandler(mockSvc)

	req, err := newMultipartRequest(t, "file", "test.xlsx", "dummy Excel content")
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.UploadEmailsFile(c)

	// Error should be attached to Gin context errors
	assert.NotEmpty(t, c.Errors)
	mockSvc.AssertExpectations(t)
}
