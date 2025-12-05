package handler

import (
	"fmt"
	"html/template"
	"li-acc/internal/model"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// MainPageData represents data for main_page
type MainPageData struct {
	Errors         []string
	ErrorMsg       string
	SuccessMsg     string
	MissingPayers  []string
	FailedEmails   []string
	SentAmount     int
	PartialSuccess bool
}

// HistoryPageData represents data for history_page
type HistoryPageData struct {
	Files []model.File
}

//
//// FileInfo stores information about file (used for file history)
//type FileInfo struct {
//	URL      string // path for download
//	Filename string
//}

// SettingsPageData represents data for settings_page
type SettingsPageData struct {
	Errors           []string
	ErrorMsg         string
	SuccessMsgEmails string
}

type UIHandler struct {
	templates map[string]*template.Template
	apiClient *APIClient
}

func NewUIHandler(apiBaseURL string, templatesPath string) *UIHandler {
	templates := make(map[string]*template.Template)

	pages := []string{"main_page", "history_page", "settings_page", "docs_page"}

	for _, page := range pages {
		tmpl := template.Must(template.ParseFiles(
			filepath.Join(templatesPath, "layouts/base.gohtml"),
			filepath.Join(templatesPath, "layouts/header.gohtml"),
			filepath.Join(templatesPath, "pages/"+page+".gohtml"),
		))
		templates[page] = tmpl
	}

	return &UIHandler{
		templates: templates,
		apiClient: NewAPIClient(apiBaseURL),
	}
}

// Главная страница - загрузка плательщиков
func (h *UIHandler) MainPage(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		data := MainPageData{}
		h.renderTemplate(c.Writer, "main_page", data)
		return
	}

	// POST - обработка загрузки
	file, err := c.FormFile("file")
	if err != nil {
		data := MainPageData{
			ErrorMsg: "Не удалось получить файл",
		}
		h.renderTemplate(c.Writer, "main_page", data)
		return
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		data := MainPageData{
			ErrorMsg: "Ошибка чтения файла",
		}
		h.renderTemplate(c.Writer, "main_page", data)
		return
	}
	defer src.Close()

	// Вызываем API
	resp, err := h.apiClient.UploadPayers(file.Filename, src)
	if err != nil {
		data := MainPageData{
			ErrorMsg: err.Error(),
		}
		h.renderTemplate(c.Writer, "main_page", data)
		return
	}

	// Формируем сообщение с учетом partial success
	var successMsg string
	if resp.PartialSuccess {
		successMsg = fmt.Sprintf("Файл обработан частично. Отправлено писем: %d", resp.SentAmount)
		if len(resp.FailedEmails) > 0 {
			successMsg += fmt.Sprintf(". Не удалось отправить %d:\n", len(resp.FailedEmails))
			successMsg += strings.Join(resp.FailedEmails, ", ")
		}
		if len(resp.MissingPayers) > 0 {
			successMsg += fmt.Sprintf(". Не найдены плательщики %d:\n", len(resp.MissingPayers))
			successMsg += strings.Join(resp.MissingPayers, ", ")
		}
	} else {
		successMsg = fmt.Sprintf("Файл успешно обработан! Отправлено писем: %d", resp.SentAmount)
	}

	data := MainPageData{
		SentAmount:     resp.SentAmount,
		PartialSuccess: resp.PartialSuccess,
		MissingPayers:  resp.MissingPayers,
		FailedEmails:   resp.FailedEmails,
		SuccessMsg:     successMsg,
	}
	h.renderTemplate(c.Writer, "main_page", data)
}

// История
func (h *UIHandler) HistoryPage(c *gin.Context) {
	files, err := h.apiClient.GetHistory()
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка получения истории: %v", err)
		return
	}

	data := HistoryPageData{
		Files: files,
	}
	h.renderTemplate(c.Writer, "history_page", data)
}

// Настройки - загрузка почт
func (h *UIHandler) SettingsPage(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		data := SettingsPageData{}
		h.renderTemplate(c.Writer, "settings_page", data)
		return
	}

	// POST - обработка загрузки
	file, err := c.FormFile("file")
	if err != nil {
		data := SettingsPageData{
			ErrorMsg: "Не удалось получить файл",
		}
		h.renderTemplate(c.Writer, "settings_page", data)
		return
	}

	src, err := file.Open()
	if err != nil {
		data := SettingsPageData{
			ErrorMsg: "Ошибка чтения файла",
		}
		h.renderTemplate(c.Writer, "settings_page", data)
		return
	}
	defer src.Close()

	// Вызываем API
	_, err = h.apiClient.UploadEmails(file.Filename, src)
	if err != nil {
		data := SettingsPageData{
			ErrorMsg: fmt.Sprintf("Ошибка обработки: %v", err),
		}
		h.renderTemplate(c.Writer, "settings_page", data)
		return
	}

	data := SettingsPageData{
		SuccessMsgEmails: "Файл с почтами успешно загружен!",
	}
	h.renderTemplate(c.Writer, "settings_page", data)
}

// Документация
func (h *UIHandler) DocsPage(c *gin.Context) {
	h.renderTemplate(c.Writer, "docs_page", nil)
}

func (h *UIHandler) renderTemplate(w http.ResponseWriter, page string, data interface{}) {
	tmpl, ok := h.templates[page]
	if !ok {
		http.Error(w, "Template not found: "+page, http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
