package handler

import (
	"fmt"
	"li-acc/internal/model"
	"li-acc/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	service service.HistoryService
}

func NewHistoryHandler(s service.HistoryService) *HistoryHandler {
	return &HistoryHandler{service: s}
}

// GetFilesHistory godoc
//
// @Summary      Retrieve file history records
// @Description  Returns a list of all file records with details including file name and raw data bytes.
// @Tags         history
// @Accept       json
// @Produce      json
// @Success      200  {object}  FilesHistoryResponse  "List of file records"
// @Failure      500  {object}  map[string]string    "Internal server error"
// @Router       /history [get]
func (h *HistoryHandler) GetFilesHistory(c *gin.Context) {
	// Call service layer to get records, propagate context
	records, err := h.service.GetRecords(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	// Form FilePath for each file
	for i := range records {
		records[i].FilePath = fmt.Sprintf("%s/%s", model.PayersXlsDir, records[i].FileName)
	}

	// Respond with JSON and 200 status
	c.JSON(http.StatusOK, FilesHistoryResponse{Files: records})
}
