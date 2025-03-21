package helpers

import (
	"context"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreatePostgresContainer(ctx context.Context, initialScriptPaths ...string) (*postgres.PostgresContainer, error) {
	const (
		image           = "postgres:17.4-alpine"
		postgreUser     = "root"
		postgrePassword = "pass"
		postgreDB       = "migrago"
	)

	return postgres.Run(ctx,
		image,
		postgres.WithDatabase(postgreDB),
		postgres.WithUsername(postgreUser),
		postgres.WithPassword(postgrePassword),
		postgres.WithInitScripts(initialScriptPaths...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
}
