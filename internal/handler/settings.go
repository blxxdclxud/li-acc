package handler

import (
	"li-acc/internal/service"
	"net/http"

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
	err := h.service.ProcessEmailsFile(c.Request.Context(), filename, fileData)
	if err != nil {
		// Attach error for centralized middleware or localize it here
		c.Error(err)
		return
	}

	// Success response
	c.JSON(http.StatusOK, EmailsFileUploadResponseSuccess{Message: "file processed successfully"})
}
