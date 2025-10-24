package handler

import (
	"errors"
	"fmt"
	"io"
	"li-acc/internal/errs"
	"li-acc/internal/service"
	"li-acc/pkg/xls"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Localizer returns a localized string message based on an error's type.
// It accepts an error, inspects the custom error type,
// and returns the corresponding translated message for a user.
// If no translation is found, or the error is of "server" type, it returns a default or fallback message.
func Localizer(err error) string {
	var codedError errs.CodedError
	if errors.As(err, &codedError) && errs.IsUserError(err) {
		//
		// ==== xls package errors ===
		//
		var me *xls.MissingEmailsError
		if errors.As(err, &me) {
			var lines []string
			for _, i := range me.MissingLines {
				lines = append(lines, strconv.Itoa(i))
			}
			return "Ошибка Excel таблицы с почтами: пропущены данные в рядах " + strings.Join(lines, ", ")
		}

		var mp *xls.MissingParamsError
		if errors.As(err, &mp) {
			// handle MissingParamsError
			return fmt.Sprintf("На листе %s отстутствуют следующие обязательные параметры: %s",
				mp.Sheet, strings.Join(mp.Missing, ", "))
		}

		var ms *xls.MissingSheetError
		if errors.As(err, &ms) {
			return "В Excel таблице отсутствует лист " + ms.Sheet
		}

		var mc *xls.MissingPayersSheetColumns
		if errors.As(err, &mc) {
			return fmt.Sprintf("На листе %s неверное число колонок: имеется %d, ожидается %d",
				xls.PayersSheet, mc.Have, mc.Want)
		}

		//
		// ==== service layer errors ===
		//
		emailMappingBaseMsg := "Некоторые плательщики не имеют сопоставленных email адресов"
		emailSendingBaseMsg := "Неудалось отправить квитанции некоторым получателям"

		var c *service.CompositeError
		if errors.As(err, &c) {
			for _, subErr := range c.Unwrap() {
				// Check for known error types per sub-error
				var es *service.EmailSendingError
				if errors.As(subErr, &es) {
					return emailSendingBaseMsg
				}
				var em *service.EmailMappingError
				if errors.As(subErr, &em) {
					return emailMappingBaseMsg
				}
				// handle other subErr types as needed
			}
		}

		var es *service.EmailSendingError
		if errors.As(err, &es) {
			return emailSendingBaseMsg
		}

		var em *service.EmailMappingError
		if errors.As(err, &em) {
			var payers []string
			for p := range em.MapPayerReceipt {
				payers = append(payers, p)
			}
			return emailMappingBaseMsg + strings.Join(payers, ", ")
		}

	}
	return "Ошибка сервера, попробуйте позже, или обратитесь к администратору сервиса"
}

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
