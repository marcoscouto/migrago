package migrago

import (
	"database/sql"
	"os"
	"regexp"

	"github.com/marcoscouto/migrago/internal/data"
	"github.com/marcoscouto/migrago/internal/processor"
	"github.com/marcoscouto/migrago/internal/transaction"
)

type Migrago interface {
	ExecuteMigrations(folderPath string) error
}

type migrago struct {
	regex            *regexp.Regexp
	uniqueMigrations map[uint64]bool
	connection       *sql.DB
}

func New(connection *sql.DB) Migrago {
	const (
		pattern = `^(\d+)_([a-zA-Z0-9_-]+)\.sql$`
	)
	regex := regexp.MustCompile(pattern)
	uniqueMigrations := make(map[uint64]bool)

	return &migrago{
		regex:            regex,
		uniqueMigrations: uniqueMigrations,
		connection:       connection,
	}
}

func (m *migrago) ExecuteMigrations(folderPath string) error {
	return transaction.New(m.connection).Execute(func(tx *sql.Tx) error {
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
				Regex:              m.regex,
				UniqueMigrations:   m.uniqueMigrations,
				ExecutedMigrations: executedMigrations,
				FileName:           f.Name(),
				FolderPath:         folderPath,
				LastMigration:      lastMigrationExecuted,
				DbTx:               tx,
			}

			executedMigrations := processor.NewExecuteMigration(data, nil)
			validateOrder := processor.NewValidateOrder(data, executedMigrations)
			verifyExecuted := processor.NewVerifyExecuted(data, validateOrder)
			validateDuplication := processor.NewValidateDuplication(data, verifyExecuted)
			getVersion := processor.NewGetVersion(data, validateDuplication)
			validatePattern := processor.NewValidatePattern(data, getVersion)

			return validatePattern.Execute()
		}
		return nil
	})
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
