package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func getExcelFileFromMultipart(c *gin.Context) (string, []byte) {
	// Parse multipart form with a reasonable max size (e.g., 10MB)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "слишком большой файл"})
		return "", nil
	}

	// Get file header
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return "", nil
	}

	// Validate file extension
	if !strings.HasSuffix(fileHeader.Filename, ".xls") &&
		!strings.HasSuffix(fileHeader.Filename, ".xlsx") &&
		!strings.HasSuffix(fileHeader.Filename, ".xlsm") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "можно загрузить только файлы Excel (.xls, .xlsx, .xlsm)"})
		return "", nil
	}

	// Open uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read uploaded file"})
		return "", nil
	}
	defer file.Close()

	// Read entire file content into memory
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file content"})
		return "", nil
	}

	return fileHeader.Filename, fileData
}
