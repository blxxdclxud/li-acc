package main

import (
	"context"
	"errors"
	"fmt"
	"li-acc/config"
	"li-acc/internal/handler"
	"li-acc/internal/model"
	"li-acc/internal/service"
	"li-acc/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const ENV = "production"

func main() {
	// load configs from .env file
	cfg := config.LoadConfig()

	if err := logger.Init(ENV); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	// ==== Setup Service Layer

	// database connection string
	db := cfg.DB
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		db.User, db.Password, db.Host, db.Port, db.DbName)

	// create service manager
	serviceManager, err := service.NewManager(dsn, cfg.ConvertAPI.PublicKey, (model.SMTP)(cfg.SMTP))
	if err != nil {
		logger.Fatal("failed to create service manager", zap.Error(err))
	}

	// ==== Setup Servers

	// API server
	apiRouter := handler.SetupRouter(serviceManager)
	apiHost := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	apiServer := &http.Server{
		Addr:    apiHost,
		Handler: apiRouter,
	}

	// Metrics server
	metricsRouter := gin.New()
	metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))
	metricsRouter.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	metricsHost := cfg.Metrics.Host + ":" + cfg.Metrics.Port
	metricsServer := &http.Server{
		Addr:    metricsHost,
		Handler: metricsRouter,
	}

	// Start API server
	go func() {
		logger.Info("API server starting", zap.String("address", apiHost))
		if err := apiServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("API server failed", zap.Error(err))
		}
	}()

	// Start metrics server
	go func() {
		logger.Info("Metrics server starting", zap.String("address", metricsHost))
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Metrics server failed", zap.Error(err))
		}
	}()

	start := time.Now()
	defer func() {
		end := time.Now()
		duration := time.Since(start)
		logger.Info("Server was available",
			zap.String("from", start.Format("15:04:05 02.01.2006")),
			zap.String("till", end.Format("15:04:05 02.01.2006")),
			zap.Duration("duration", duration),
		)
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Shutting down servers...")

	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error("API server shutdown error", zap.Error(err))
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("Metrics server shutdown error", zap.Error(err))
	}

	logger.Info("Servers stopped gracefully")
}
