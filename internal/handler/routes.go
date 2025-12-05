package handler

import (
	"li-acc/internal/middleware"
	"li-acc/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ApiRoutesGroup          = "/api"
	ApiEndpointUploadPayers = "/upload-payers"
	ApiEndpointUploadEmails = "/settings/upload-emails"
	ApiEndpointGetHistory   = "/history"
)

func SetupRouter(manager service.ManagerIface, uiHandler *UIHandler) *gin.Engine {
	r := gin.Default()

	// Add global middleware here (logging, error handling, CORS, etc.)
	r.Use(gin.Recovery(), middleware.LoggingMiddleware(middleware.LoggerConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     10 << 40,
	}), middleware.ErrorHandler())

	// Create handler instances
	mainHandler := NewMainHandler(manager)
	settingsHandler := NewSettingsHandler(manager.SettingsService())
	historyHandler := NewHistoryHandler(manager.HistoryService())

	// === API Groups ===
	api := r.Group("/api")
	{
		// Upload payers Excel file
		api.POST(ApiEndpointUploadPayers, mainHandler.UploadPayersFile)

		// Upload settings or sender emails file
		api.POST(ApiEndpointUploadEmails, settingsHandler.UploadEmailsFile)

		// Get history of uploaded files
		api.GET(ApiEndpointGetHistory, historyHandler.GetFilesHistory)
	}

	// === Static files ===
	r.Static("/static", "./static")
	r.Static("/tmp", "./tmp")

	// === UI routes ===
	r.GET("/", uiHandler.MainPage)
	r.POST("/", uiHandler.MainPage)

	r.GET("/history", uiHandler.HistoryPage)

	r.GET("/settings", uiHandler.SettingsPage)
	r.POST("/settings", uiHandler.SettingsPage)

	r.GET("/documentation", uiHandler.DocsPage)

	// === Health-check route ===
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
