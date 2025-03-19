package migrago

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type Migrago interface {
	ExecuteMigrations(folderPath string, connection *sql.DB) error
}

type migrago struct {
	regex            *regexp.Regexp
	uniqueMigrations map[uint64]bool
}

func New() Migrago {
	regex := regexp.MustCompile(`^(\d+)_([a-zA-Z0-9_-]+)\.sql$`)
	uniqueMigrations := make(map[uint64]bool)
	return &migrago{
		regex:            regex,
		uniqueMigrations: uniqueMigrations,
	}
}

func (m *migrago) ExecuteMigrations(folderPath string, connection *sql.DB) error {
	transaction, err := connection.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	executedMigrations, lastMigrationExecuted, err := getExecutedMigrations(transaction)
	if err != nil {
		return err
	}

	fileNames, err := readFilesNames(folderPath)
	if err != nil {
		return err
	}

	for _, f := range fileNames {
		if err := validateMigrationPattern(f, m.regex); err != nil {
			return err
		}

		version, err := getMigrationVersion(f, m.regex)
		if err != nil {
			return err
		}

		if err := validateMigrationDuplication(version, m.uniqueMigrations); err != nil {
			return err
		}

		if _, ok := executedMigrations[version]; ok {
			continue
		}

		if err := validateMigrationOrder(version, &lastMigrationExecuted); err != nil {
			return err
		}

		if err := executeMigrations(transaction, version, folderPath, f); err != nil {
			return err
		}

	}
	return nil
}

func readFilesNames(folderPath string) (names []string, err error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		names = append(names, f.Name())
	}

	return names, nil
}

func getExecutedMigrations(transaction *sql.Tx) (executedMigrations map[uint64]Migration, lastMigrationExecuted uint64, err error) {
	executedMigrations = make(map[uint64]Migration)
	lastMigrationExecuted = 0

	result, err := transaction.Query("SELECT version, name, checksum, applied_at FROM migrago ORDER BY version DESC")
	if err != nil {
		return nil, 0, err
	}
	defer result.Close()

	for result.Next() {
		var migration Migration
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

func validateMigrationPattern(fileName string, regex *regexp.Regexp) error {
	match := regex.Match([]byte(fileName))
	if !match {
		return fmt.Errorf("invalid migration filename format")
	}
	return nil
}

func validateMigrationDuplication(version uint64, uniqueMigrations map[uint64]bool) error {
	if uniqueMigrations[version] {
		return fmt.Errorf("duplicated migration file")
	}
	uniqueMigrations[version] = true
	return nil
}

func getMigrationVersion(fileName string, regex *regexp.Regexp) (uint64, error) {
	matches := regex.FindStringSubmatch(fileName)
	version, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func validateMigrationOrder(version uint64, lastMigrationExecuted *uint64) error {
	*lastMigrationExecuted++
	if version != *lastMigrationExecuted {
		return fmt.Errorf("the migration is out of order")
	}
	return nil
}

func executeMigrations(transaction *sql.Tx, version uint64, migrationsPath, fileName string) error {
	path := filepath.Join(migrationsPath, fileName)
	byteContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(byteContent)
	if _, err := transaction.Exec(content); err != nil {
		return err
	}

	s := sha256.Sum256(byteContent)
	checksum := fmt.Sprintf("%x", s)

	if _, err := transaction.Exec("INSERT INTO migrago (version, name, checksum, applied_at) VALUES ($1, $2, $3, $4)", version, fileName, checksum, time.Now().UTC()); err != nil {
		return err
	}

	return nil
}
