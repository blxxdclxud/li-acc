package integration

import (
	"context"
	"fmt"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(ctx context.Context) (string, func(), error) {
	user := "test"
	pass := "test"
	db := "test"

	req := tc.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432"},
		Env: map[string]string{
			"POSTGRES_USER":     user,
			"POSTGRES_DB":       db,
			"POSTGRES_PASSWORD": pass,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", func() {}, err
	}

	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port.Port(), db)

	cleanup := func() { c.Terminate(ctx) }

	return dsn, cleanup, nil
}
