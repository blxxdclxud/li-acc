package handler_test

import (
	"encoding/json"
	"errors"
	"li-acc/internal/handler"
	"li-acc/internal/mocks"
	"li-acc/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetFilesHistory_Success(t *testing.T) {
	mockService := new(mocks.HistoryService)
	mockService.On("GetRecords", mock.Anything).Return([]model.File{
		{FileName: "test.txt", FileData: []byte("abc")},
	}, nil)

	h := handler.NewHistoryHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Attach request context
	c.Request = httptest.NewRequest(http.MethodGet, "/history", nil)

	h.GetFilesHistory(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response handler.FilesHistoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Files, 1)
	assert.Equal(t, "test.txt", response.Files[0].FileName)
}

func TestGetFilesHistory_Error(t *testing.T) {
	mockService := new(mocks.HistoryService)
	mockService.On("GetRecords", mock.Anything).Return([]model.File{}, errors.New("db error"))

	handler := handler.NewHistoryHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/history", nil)

	handler.GetFilesHistory(c)

	// You can check if error was attached to context etc.
	assert.NotEmpty(t, c.Errors)
}
