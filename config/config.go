package config

import (
	"fmt"
	"li-acc/pkg/logger"
	"os"
	"path/filepath"
	"sync"

	"github.com/gabrielsoaressantos/env/v8"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Host string `env:"SERVER_HOST,notEmpty"`
		Port int    `env:"SERVER_PORT,notEmpty"`
	}

	DB struct {
		Host     string `env:"POSTGRES_HOST,notEmpty"`
		Port     int    `env:"POSTGRES_PORT,notEmpty"`
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

const FilePath = "./.env"

var (
	cfg  *Config
	once sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		absFP, _ := filepath.Abs(FilePath)
		_, err := os.ReadFile(FilePath)
		if err != nil {
			logger.Fatal(fmt.Sprintf("error finding `%s` file", absFP), zap.Error(err))
		}

		cfg = &Config{}
		err = env.ParseNested(cfg)
		if err != nil {
			logger.Fatal(fmt.Sprintf("error parsing config file `%s`", absFP), zap.Error(err))
		}
	})

	return cfg
}
