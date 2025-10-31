package handler_test

import (
	"encoding/json"
	"errors"
	"li-acc/internal/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"li-acc/internal/handler"
	"li-acc/internal/service"
)

func TestUploadPayersFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.Manager)
	respMap := map[string]string{"a@a.com": "/path/to/a.pdf"}
	svc.On("ProcessPayersFile", mock.Anything, "test.xlsx", mock.Anything).Return(respMap, 1, nil)

	h := handler.NewMainHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = newMultipartRequest(t)

	h.UploadPayersFile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handler.PayersFileUploadResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "file processed successfully", resp.Message)
	assert.Equal(t, 1, resp.SentAmount)
	assert.False(t, resp.PartialSuccess)
	assert.Empty(t, resp.FailedEmails)
	assert.Empty(t, resp.MissingPayers)

	svc.AssertExpectations(t)
}

func TestUploadPayersFile_PartialSuccess_EmailSendingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.Manager)

	failedEmails := map[string]string{"fail1@example.com": "smtp error"}
	compositeErr := &service.CompositeError{
		Errors: []error{
			&service.EmailSendingError{MapReceiverCause: failedEmails},
		},
	}

	respMap := map[string]string{"success@example.com": "/path/to/success.pdf"}
	svc.On("ProcessPayersFile", mock.Anything, "test.xlsx", mock.Anything).
		Return(respMap, 1, compositeErr)

	h := handler.NewMainHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = newMultipartRequest(t)

	h.UploadPayersFile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handler.PayersFileUploadResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.PartialSuccess)
	assert.Equal(t, []string{"fail1@example.com"}, resp.FailedEmails)
	assert.Empty(t, resp.MissingPayers)
	assert.Equal(t, 1, resp.SentAmount)

	svc.AssertExpectations(t)
}

func TestUploadPayersFile_PartialSuccess_EmailMappingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.Manager)

	missingPayers := map[string]string{"payer1": "/receipt/path"}
	compositeErr := &service.CompositeError{
		Errors: []error{
			&service.EmailMappingError{MapPayerReceipt: missingPayers},
		},
	}

	respMap := map[string]string{"success@example.com": "/path/to/success.pdf"}
	svc.On("ProcessPayersFile", mock.Anything, "test.xlsx", mock.Anything).
		Return(respMap, 1, compositeErr)

	h := handler.NewMainHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = newMultipartRequest(t)

	h.UploadPayersFile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp handler.PayersFileUploadResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.PartialSuccess)
	assert.Empty(t, resp.FailedEmails)
	assert.Equal(t, []string{"payer1"}, resp.MissingPayers)
	assert.Equal(t, 1, resp.SentAmount)

	svc.AssertExpectations(t)
}

func TestUploadPayersFile_FullFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.Manager)

	svc.On("ProcessPayersFile", mock.Anything, "test.xlsx", mock.Anything).
		Return(make(map[string]string), 0, errors.New("unexpected error"))

	h := handler.NewMainHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = newMultipartRequest(t)

	h.UploadPayersFile(c)

	// Full failure attaches error to context and does not write successful JSON response
	assert.NotEmpty(t, c.Errors)

	svc.AssertExpectations(t)
}
