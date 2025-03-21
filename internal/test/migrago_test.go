package test

// https://golang.testcontainers.org/quickstart/

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/marcoscouto/migrago"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MigrationProcessorTestSuite struct {
	suite.Suite
	ctx       context.Context
	container *postgres.PostgresContainer
	db        *sql.DB
	migrago   migrago.Migrago
}

func (suite *MigrationProcessorTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	container, err := createPostgresContainer(suite.ctx)
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.container = container
}

func (suite *MigrationProcessorTestSuite) SetupTest() {
	const (
		database = "postgres"
	)

	connString, err := suite.container.ConnectionString(suite.ctx)
	if err != nil {
		suite.T().Fatal(err)
	}

	db, err := sql.Open(database, fmt.Sprint(connString, "sslmode=disable"))
	if err != nil {
		suite.T().Fatal(err)
	}

	if err := db.Ping(); err != nil {
		suite.T().Fatal(err)
	}

	suite.db = db
	suite.migrago = migrago.New(db)
}

func (suite *MigrationProcessorTestSuite) TearDownTest() {
	suite.db.Close()
}

func (suite *MigrationProcessorTestSuite) TearDownSuite() {
	suite.container.Terminate(suite.ctx)
}

func TestMigrationProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationProcessorTestSuite))
}

func (suite *MigrationProcessorTestSuite) TestExecuteSuccessfully() {
	err := suite.migrago.ExecuteMigrations("valid")
	suite.NoError(err)
}

func (suite *MigrationProcessorTestSuite) TestExecutePatternError() {
	err := suite.migrago.ExecuteMigrations("invalid_pattern")
	suite.Error(err)
	suite.Equal("invalid migration filename format", err.Error())
}

func (suite *MigrationProcessorTestSuite) TestExecuteDuplicatedError() {
	err := suite.migrago.ExecuteMigrations("duplicated")
	suite.Error(err)
	suite.Equal("duplicated migration file", err.Error())
}

func (suite *MigrationProcessorTestSuite) TestExecuteOutOfOrderError() {
	err := suite.migrago.ExecuteMigrations("out_of_order")
	suite.Error(err)
	suite.Equal("the migration is out of order", err.Error())
}

func createPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
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
		postgres.WithInitScripts(filepath.Join("init.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
}
