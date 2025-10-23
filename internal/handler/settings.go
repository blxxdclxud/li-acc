package handler

import (
	"io"
	"li-acc/internal/service"
	"net/http"
	"strings"

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
// @Success      200  {object}  FileUploadResponseSuccess "File processed successfully"
// @Failure      400  {object}  map[string]string        "Bad request (invalid file, missing, or too large)"
// @Failure      500  {object}  map[string]string        "Internal server error"
//
// @Router       /settings [post]
func (h *SettingsHandler) UploadEmailsFile(c *gin.Context) {
	// Parse multipart form with a reasonable max size (e.g., 10MB)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		// handle error, respond bad request
		c.JSON(http.StatusBadRequest, gin.H{"error": "слишком большой файл"})
		return
	}

	// Get file header
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Validate file extension
	if !(strings.HasSuffix(fileHeader.Filename, ".xls") ||
		strings.HasSuffix(fileHeader.Filename, ".xlsx") ||
		strings.HasSuffix(fileHeader.Filename, ".xlsm")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "можно загрузить только файлы Excel (.xls, .xlsx, .xlsm)"})
		return
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read uploaded file"})
		return
	}
	defer file.Close()

	// Read entire file content into memory
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file content"})
		return
	}

	// Call service ProcessEmailsFile with context, filename, and file data
	err = h.service.ProcessEmailsFile(c.Request.Context(), fileHeader.Filename, fileData)
	if err != nil {
		// Attach error for centralized middleware or localize it here
		c.Error(err)
		return
	}

	// Success response
	c.JSON(http.StatusOK, FileUploadResponseSuccess{Message: "file processed successfully"})
}
