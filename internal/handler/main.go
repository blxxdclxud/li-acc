package handler

import (
	"errors"
	"li-acc/internal/metrics"
	"li-acc/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MainHandler struct {
	service service.ManagerIface
}

func NewMainHandler(s service.ManagerIface) *MainHandler {
	return &MainHandler{service: s}
}

// UploadPayersFile godoc
//
// @Summary      Upload and process an Excel file containing payer emails
// @Description  Accepts a multipart/form-data POST request with an Excel file (.xls, .xlsx, .xlsm).
//
//	Parses payers, generates receipts, and sends emails.
//	Supports partial success: includes counts of sent emails and lists of failed emails or missing payers.
//	Returns detailed JSON response indicating overall success and any partial failures.
//
// @Tags         settings
// @Accept       multipart/form-data
// @Produce      json
//
// @Param        file  formData  file  true  "Excel file to upload. Allowed extensions: .xls, .xlsx, .xlsm"
//
// @Success      200  {object}  PayersFileUploadResponse  "File processed successfully with optional partial failure details"
// @Failure      400  {object}  map[string]string        "Bad request errors (file missing, invalid file type, too large)"
// @Failure      500  {object}  map[string]string        "Internal server errors"
//
// @Router       /upload-payers [post]
func (h *MainHandler) UploadPayersFile(c *gin.Context) {
	filename, fileData := getExcelFileFromMultipart(c)
	if filename == "" || fileData == nil {
		return // error response already sent inside the function
	}

	// Call service ProcessPayersFile with context, filename, and file data
	start := time.Now()
	_, sentCount, err := h.service.ProcessPayersFile(c.Request.Context(), filename, fileData)

	// update file processing latency metric
	duration := time.Since(start).Seconds()
	if err != nil {
		metrics.PayersFileLatency.WithLabelValues("failure").Observe(duration)
	} else {
		metrics.PayersFileLatency.WithLabelValues("success").Observe(duration)
	}

	response := PayersFileUploadResponse{
		Message:    "file processed successfully",
		SentAmount: sentCount,
	}

	if err != nil {
		var errorStage string

		var compositeErr *service.CompositeError
		if errors.As(err, &compositeErr) {
			response.PartialSuccess = true

			// For partial failures, error_stage is "send_mails" or "email_mapping"
			for _, e := range compositeErr.Errors {
				switch typedErr := e.(type) {
				case *service.EmailSendingError:
					errorStage = "send_mails"
					var failed []string
					for email := range typedErr.MapReceiverCause {
						failed = append(failed, email)
					}
					response.FailedEmails = failed
				case *service.EmailMappingError:
					errorStage = "email_mapping"
					var missed []string
					for email := range typedErr.MapPayerReceipt {
						missed = append(missed, email)
					}
					response.MissingPayers = missed
				}
			}

			metrics.FileProcessedTotal.WithLabelValues("failure", errorStage, "payers").Inc()
		} else {
			metrics.FileProcessedTotal.WithLabelValues("failure", "not-partial", "payers").Inc()
			// Handle a full failure (system/user error that is not partial)
			c.Error(err)
			return
		}
	}
	metrics.FileProcessedTotal.WithLabelValues("success", "", "payers").Inc()

	c.JSON(http.StatusOK, response)
}
