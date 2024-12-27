package main

import (
	"fmt"
	"go.uber.org/zap"
	"li-acc/internal/db"
	"li-acc/internal/repository"
	"li-acc/pkg/logger"
)

func main() {
	err := logger.Init("development")
	if err != nil {
		zap.L().Fatal("cannot initialize logger: ", zap.Error(err))
	}

	repo, err := repository.ConnectDB("postgres://admin:liacc@127.0.0.1:5432/accounting?sslmode=disable")
	if err != nil {
		fmt.Println(err)
		return
	}

	changed, err := db.ApplyMigrations(repo.DB, "/internal/db/migrations")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(changed)

}
