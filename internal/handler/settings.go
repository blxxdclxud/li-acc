package handler

import (
	"li-acc/internal/metrics"
	"li-acc/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service service.SettingsService
}

func NewSettingsHandler(s service.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		service: s,
	}
}

// UploadEmailsFile godoc
//
// @Summary      Upload and process an Excel file containing payer emails
// @Description  Accepts a multipart/form-data POST request containing an Excel file (.xls, .xlsx, .xlsm).
//
//	Validates file type and size. Parses, validates, and stores payer emails.
//	Returns success message or relevant error messages in JSON format.
//
// @Tags         settings
// @Accept       multipart/form-data
// @Produce      json
//
// @Param        file formData file true "Excel file to upload. Allowed extensions: .xls, .xlsx, .xlsm"
//
// @Success      200  {object}  EmailsFileUploadResponseSuccess "File processed successfully"
// @Failure      400  {object}  map[string]string        "Bad request (invalid file, missing, or too large)"
// @Failure      500  {object}  map[string]string        "Internal server error"
//
// @Router       /settings/upload-emails [post]
func (h *SettingsHandler) UploadEmailsFile(c *gin.Context) {
	filename, fileData := getExcelFileFromMultipart(c)
	if filename == "" || fileData == nil {
		return // error response already sent inside the function
	}
	// Call service ProcessEmailsFile with context, filename, and file data
	start := time.Now()
	err := h.service.ProcessEmailsFile(c.Request.Context(), filename, fileData)

	// update file processing latency metric
	duration := time.Since(start).Seconds()
	if err != nil {
		metrics.EmailsFileLatency.WithLabelValues("failure").Observe(duration)
	} else {
		metrics.EmailsFileLatency.WithLabelValues("success").Observe(duration)
	}

	if err != nil {
		metrics.FileProcessedTotal.WithLabelValues("failure", "", "emails").Inc()

		// Attach error for centralized middleware or localize it here
		c.Error(err)
		return
	}

	metrics.FileProcessedTotal.WithLabelValues("success", "", "emails").Inc()

	// Success response
	c.JSON(http.StatusOK, EmailsFileUploadResponseSuccess{Message: "file processed successfully"})
}
