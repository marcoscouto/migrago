package test

// https://golang.testcontainers.org/quickstart/

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/marcoscouto/migrago"
	"github.com/marcoscouto/migrago/internal/errors"
	"github.com/marcoscouto/migrago/test/helpers"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	validMigrationsDir          = "migrations/valid"
	invalidPatternMigrationsDir = "migrations/invalid_pattern"
	duplicatedMigrationsDir     = "migrations/duplicated"
	outOfOrderMigrationsDir     = "migrations/out_of_order"
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
	container, err := helpers.CreatePostgresContainer(suite.ctx, filepath.Join("scripts", "init.sql"))
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.container = container
}

func (suite *MigrationProcessorTestSuite) SetupTest() {
	const (
		database    = "postgres"
		disabledSSL = "sslmode=disable"
	)

	connString, err := suite.container.ConnectionString(suite.ctx)
	if err != nil {
		suite.T().Fatal(err)
	}

	db, err := sql.Open(database, fmt.Sprint(connString, disabledSSL))
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
	err := suite.migrago.ExecuteMigrations(validMigrationsDir)
	suite.NoError(err)
}

func (suite *MigrationProcessorTestSuite) TestExecutePatternError() {
	err := suite.migrago.ExecuteMigrations(invalidPatternMigrationsDir)
	suite.Error(err)
	suite.Equal(errors.ErrInvalidPattern, err)
}

func (suite *MigrationProcessorTestSuite) TestExecuteDuplicatedError() {
	err := suite.migrago.ExecuteMigrations(duplicatedMigrationsDir)
	suite.Error(err)
	suite.Equal(errors.ErrDuplicatedFile, err)
}

func (suite *MigrationProcessorTestSuite) TestExecuteOutOfOrderError() {
	err := suite.migrago.ExecuteMigrations(outOfOrderMigrationsDir)
	suite.Error(err)
	suite.Equal(errors.ErrOutOfOrder, err)
}
