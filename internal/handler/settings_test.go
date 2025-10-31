package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"li-acc/internal/handler"
	"li-acc/internal/mocks"
)

// helper to create multipart file upload request
func newMultipartRequest(t *testing.T) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.xlsx")
	assert.NoError(t, err)
	_, err = part.Write([]byte("dummy content"))
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadEmailsFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.SettingsService)
	// Expect ProcessEmailsFile called with any context, filename, and byte slice, returning nil error
	mockSvc.On("ProcessEmailsFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

	h := handler.NewSettingsHandler(mockSvc)

	// Prepare multipart request with dummy Excel file content
	req := newMultipartRequest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Call handler
	h.UploadEmailsFile(c)

	// Assert success response
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
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

	req := newMultipartRequest(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.UploadEmailsFile(c)

	// Error should be attached to Gin context errors
	assert.NotEmpty(t, c.Errors)
	mockSvc.AssertExpectations(t)
}
