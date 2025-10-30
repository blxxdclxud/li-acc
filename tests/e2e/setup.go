package e2e

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"li-acc/config"
	"li-acc/internal/handler"
	"li-acc/internal/model"
	migrator "li-acc/internal/repository/db"
	"li-acc/internal/service"
	"li-acc/pkg/logger"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

type TestEnvironment struct {
	PostgresContainer tc.Container
	MailHogContainer  tc.Container
	AppURL            string
	MetricsURL        string
	MailHogAPIURL     string
	MailHogSMTPPort   int
}

const MigrationsDir = "../../internal/repository/db/migrations"

func SetupTestEnvironment(t *testing.T) (*TestEnvironment, func()) {
	model.BlankReceiptPath = "../." + model.BlankReceiptPath // properly use global relative paths

	if err := logger.Init("test"); err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	ctx := context.Background()

	cfg := config.LoadConfig("test")

	// Setup PostgreSQL container
	pgContainer := setupPostgresContainer(t, ctx, cfg)

	// Setup MailHog container
	mailHogContainer, mailHogAPIURL, smtpPort := setupMailHogContainer(t, ctx)

	// Override SMTP config to use MailHog
	mailHogHost, _ := mailHogContainer.Host(ctx)
	cfg.SMTP.Host = mailHogHost
	cfg.SMTP.Port = smtpPort

	t.Logf("MailHog SMTP: %s:%d", mailHogHost, smtpPort)
	t.Logf("MailHog Web UI: %s", mailHogAPIURL)

	// Start application
	app := startTestApp(t, cfg)

	env := &TestEnvironment{
		PostgresContainer: pgContainer,
		MailHogContainer:  mailHogContainer,
		AppURL:            "http://" + cfg.Server.Host + ":" + cfg.Server.Port,
		MetricsURL:        "http://" + cfg.Metrics.Host + ":" + cfg.Metrics.Port,
		MailHogAPIURL:     mailHogAPIURL,
		MailHogSMTPPort:   smtpPort,
	}

	// Cleanup function
	cleanup := func() {
		app.Shutdown()

		if err := mailHogContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate MailHog: %v", err)
		}

		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate PostgreSQL: %v", err)
		}
	}

	return env, cleanup
}

// setupPostgresContainer creates and starts a PostgreSQL container
func setupPostgresContainer(t *testing.T, ctx context.Context, cfg *config.Config) tc.Container {
	user := cfg.DB.User
	pass := cfg.DB.Password
	dbname := cfg.DB.DbName
	port := cfg.DB.Port

	req := tc.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{port + ":" + port},
		Env: map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_DB":       dbname,
			"POSTGRES_PASSWORD": pass,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	pgContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start PostgreSQL container")

	host, _ := pgContainer.Host(ctx)

	t.Logf("PostgreSQL container started: %s:%s", host, port)

	return pgContainer
}

// setupMailHogContainer creates and starts a MailHog container
func setupMailHogContainer(t *testing.T, ctx context.Context) (tc.Container, string, int) {
	req := tc.ContainerRequest{
		Image:        "mailhog/mailhog:latest",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor: wait.ForHTTP("/").
			WithPort("8025/tcp").
			WithStartupTimeout(30 * time.Second),
	}

	mailHogContainer, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Failed to start MailHog container")

	// Get MailHog SMTP port (1025)
	smtpPort, err := mailHogContainer.MappedPort(ctx, "1025")
	require.NoError(t, err, "Failed to get MailHog SMTP port")

	// Get MailHog HTTP API port (8025)
	httpPort, err := mailHogContainer.MappedPort(ctx, "8025")
	require.NoError(t, err, "Failed to get MailHog HTTP port")

	host, err := mailHogContainer.Host(ctx)
	require.NoError(t, err, "Failed to get MailHog host")

	apiURL := fmt.Sprintf("http://%s:%s", host, httpPort.Port())

	t.Logf("MailHog container started - SMTP: %s:%s, API: %s", host, smtpPort.Port(), apiURL)

	p, _ := strconv.Atoi(smtpPort.Port())

	return mailHogContainer, apiURL, p
}

type App struct {
	apiServer     *http.Server
	metricsServer *http.Server
}

func (a *App) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Shutting down servers...")

	if err := a.apiServer.Shutdown(ctx); err != nil {
		logger.Error("API server shutdown error", zap.Error(err))
	}

	if err := a.metricsServer.Shutdown(ctx); err != nil {
		logger.Error("Metrics server shutdown error", zap.Error(err))
	}

	logger.Info("Servers stopped gracefully")
}

func startTestApp(t *testing.T, cfg *config.Config) *App {
	// ==== Setup Service Layer ====

	dbcfg := cfg.DB
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbcfg.User, dbcfg.Password, dbcfg.Host, dbcfg.Port, dbcfg.DbName)

	// Use sql.DB for migrations
	migrationDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("failed to open DB for migrations:", err)
	}

	applied, err := migrator.ApplyMigrations(migrationDB, MigrationsDir)
	if err != nil {
		migrationDB.Close()
		t.Fatal("failed to apply migrations:", err)
	}

	t.Logf("Migrations applied: %v", applied)

	migrationDB.Close() // <-- Closes immediately, no blocking

	// Create service manager with MailHog SMTP config
	smtp := model.SMTP{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Email:    cfg.SMTP.Email,
		Password: cfg.SMTP.Password,
		UseTLS:   false,
	}
	serviceManager, err := service.NewManager(dsn, cfg.ConvertAPI.PublicKey, smtp)
	if err != nil {
		t.Fatal("failed to create service manager", zap.Error(err))
	}

	serviceManager.SetPdfFontPath("../../static/fonts/Arial.ttf")

	// ==== Setup Servers ====

	apiRouter := handler.SetupRouter(serviceManager)
	apiHost := cfg.Server.Host + ":" + cfg.Server.Port
	apiServer := &http.Server{
		Addr:    apiHost,
		Handler: apiRouter,
	}

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

	// Start servers in goroutines
	errChan := make(chan error, 2)

	// Start API server
	go func() {
		t.Logf("API server starting on %s", apiHost)
		if err := apiServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("API server failed: %w", err)
		}
	}()

	// Start metrics server
	go func() {
		t.Logf("Metrics server starting on %s", metricsHost)
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("metrics server failed: %w", err)
		}
	}()

	// Wait for servers to be ready with health checks
	waitForServer(t, "http://"+apiHost+"/healthz", 10*time.Second)
	waitForServer(t, "http://"+metricsHost+"/healthz", 10*time.Second)

	t.Log("Test servers are ready")

	// Check if any server failed to start
	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
		// Servers started successfully
	}

	return &App{
		apiServer:     apiServer,
		metricsServer: metricsServer,
	}
}

// waitForServer polls a health check endpoint until it responds or times out
func waitForServer(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 100 * time.Millisecond}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("Server ready at %s", url)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Server at %s did not become ready within %v", url, timeout)
}
