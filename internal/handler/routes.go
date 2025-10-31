package handler

import (
	"li-acc/internal/middleware"
	"li-acc/internal/service"

	"github.com/gin-gonic/gin"
)

func SetupRouter(manager service.ManagerIface) *gin.Engine {
	r := gin.Default()

	// Add global middleware here (logging, error handling, CORS, etc.)
	r.Use(gin.Recovery(), middleware.LoggingMiddleware(middleware.LoggerConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     10 << 10,
	}), middleware.ErrorHandler())

	// Create handler instances
	mainHandler := NewMainHandler(manager)
	settingsHandler := NewSettingsHandler(manager.SettingsService())
	historyHandler := NewHistoryHandler(manager.HistoryService())

	// === API Groups ===
	api := r.Group("/api")
	{
		// Upload payers Excel file
		api.POST("/upload-payers", mainHandler.UploadPayersFile)

		// Upload settings or sender emails file
		api.POST("/settings/upload-emails", settingsHandler.UploadEmailsFile)

		// Get history of uploaded files
		api.GET("/history", historyHandler.GetFilesHistory)
	}

	// Health-check route (for infrastructure or CI)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
