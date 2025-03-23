package migrago

import (
	"database/sql"
	"os"
	"regexp"

	"github.com/marcoscouto/migrago/internal/config"
	"github.com/marcoscouto/migrago/internal/data"
	"github.com/marcoscouto/migrago/internal/processor"
	"github.com/marcoscouto/migrago/internal/transaction"
	"github.com/marcoscouto/migrago/internal/utils"
)

type Migrago interface {
	ExecuteMigrations(folderPath string) error
}

type migrago struct {
	connection       *sql.DB
	uniqueMigrations map[uint64]bool
	defaultRegex     *regexp.Regexp
	databaseUtils utils.DatabaseUtils
}

func New(connection *sql.DB, driverName string) Migrago {
	config := config.DefaultConfig()
	regex := regexp.MustCompile(config.MigrationPattern)
	uniqueMigrations := make(map[uint64]bool)
	databaseUtils := utils.NewDatabaseUtils(driverName)

	return &migrago{
		defaultRegex:     regex,
		uniqueMigrations: uniqueMigrations,
		connection:       connection,
		databaseUtils: databaseUtils,
	}

}

func (m *migrago) ExecuteMigrations(folderPath string) error {
	return transaction.New(m.connection).Execute(func(tx *sql.Tx) error {
		if err := createMigrationTable(tx); err != nil {
			return err
		}

		executedMigrations, lastMigrationExecuted, err := getExecutedMigrations(tx)
		if err != nil {
			return err
		}

		files, err := os.ReadDir(folderPath)
		if err != nil {
			return err
		}

		for _, f := range files {
			data := &data.MigrationProcessorData{
				Regex:              m.defaultRegex,
				UniqueMigrations:   m.uniqueMigrations,
				ExecutedMigrations: executedMigrations,
				FileName:           f.Name(),
				FolderPath:         folderPath,
				LastMigration:      &lastMigrationExecuted,
				DbTx:               tx,
			}

			executedMigrations := processor.NewExecuteMigration(data, nil, m.databaseUtils)
			validateOrder := processor.NewValidateOrder(data, executedMigrations)
			verifyExecuted := processor.NewVerifyExecuted(data, validateOrder)
			validateDuplication := processor.NewValidateDuplication(data, verifyExecuted)
			getVersion := processor.NewGetVersion(data, validateDuplication)
			validatePattern := processor.NewValidatePattern(data, getVersion)

			if err := validatePattern.Execute(); err != nil {
				return err
			}
		}
		return nil
	})
}

func createMigrationTable(transaction *sql.Tx) error {
	const (
		query = `CREATE TABLE IF NOT EXISTS migrago (
		version BIGINT PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		checksum VARCHAR(64) NOT NULL,
		applied_at TIMESTAMP)`
	)

	_, err := transaction.Exec(query)
	return err
}

func getExecutedMigrations(transaction *sql.Tx) (executedMigrations map[uint64]data.Migration, lastMigrationExecuted uint64, err error) {
	const (
		query = `SELECT version, name, checksum, applied_at FROM migrago ORDER BY version DESC`
	)

	executedMigrations = make(map[uint64]data.Migration)
	lastMigrationExecuted = 0

	result, err := transaction.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer result.Close()

	for result.Next() {
		var migration data.Migration
		err = result.Scan(&migration.Version, &migration.Name, &migration.Checksum, &migration.AppliedAt)
		if err != nil {
			return nil, 0, err
		}

		if migration.Version > lastMigrationExecuted {
			lastMigrationExecuted = migration.Version
		}

		executedMigrations[migration.Version] = migration
	}

	if err := result.Err(); err != nil {
		return nil, 0, err
	}

	return executedMigrations, lastMigrationExecuted, nil
}
