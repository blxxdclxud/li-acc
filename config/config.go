package config

import (
	"fmt"
	"li-acc/pkg/logger"
	"os"
	"path/filepath"
	"sync"

	"github.com/gabrielsoaressantos/env/v8"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Host string `env:"SERVER_HOST,notEmpty"`
		Port string `env:"SERVER_PORT,notEmpty"`
	}

	Metrics struct {
		Host string `env:"METRICS_HOST,notEmpty"`
		Port string `env:"METRICS_PORT,notEmpty"`
	}

	DB struct {
		Host     string `env:"POSTGRES_HOST,notEmpty"`
		Port     string `env:"POSTGRES_PORT,notEmpty"`
		User     string `env:"POSTGRES_USER,notEmpty"`
		Password string `env:"POSTGRES_PASSWORD,notEmpty"`
		DbName   string `env:"POSTGRES_DB,notEmpty"`
	}

	ConvertAPI struct {
		PublicKey  string `env:"CONVERT_API_PUBLIC_KEY,notEmpty"`
		PrivateKey string `env:"CONVERT_API_SECRET_KEY,notEmpty"`
	}

	SMTP struct {
		Host     string `env:"SMTP_HOST,notEmpty"`
		Port     int    `env:"SMTP_PORT,notEmpty"`
		Email    string `env:"SMTP_EMAIL,notEmpty"`
		Password string `env:"SMTP_PASSWORD,notEmpty"`
	}
}

const ProdFilePath = "./.env"
const TestFilePath = "./.env.test" // tests/e2e/.env.test

var (
	cfg  *Config
	once sync.Once
)

// LoadConfig loads variables from .env file.
// Passed `env` parameter = "test"|"prod"
func LoadConfig(environ string) *Config {
	once.Do(func() {
		var filePath string
		if environ == "test" {
			filePath = TestFilePath
		} else {
			filePath = ProdFilePath
		}

		absFP, _ := filepath.Abs(filePath)
		_, err := os.ReadFile(filePath)
		if err != nil {
			logger.Fatal(fmt.Sprintf("error finding `%s` file", absFP), zap.Error(err))
		}

		err = godotenv.Load(filePath)
		if err != nil {
			logger.Fatal(fmt.Sprintf("error loading env file `%s`", absFP), zap.Error(err))
		}

		cfg = &Config{}
		err = env.ParseNested(cfg)
		if err != nil {
			logger.Fatal(fmt.Sprintf("error parsing config file `%s`", absFP), zap.Error(err))
		}
	})

	return cfg
}
