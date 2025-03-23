package processor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marcoscouto/migrago/internal/data"
	"github.com/marcoscouto/migrago/internal/utils"
)

type ExecuteMigration struct {
	BaseProcessor
	databaseUtils utils.DatabaseUtils
}

func NewExecuteMigration(data *data.MigrationProcessorData, next MigrationProcessor, databaseUtils utils.DatabaseUtils) MigrationProcessor {
	return &ExecuteMigration{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
		databaseUtils: databaseUtils,
	}
}

func (v *ExecuteMigration) Execute() error {
	path := filepath.Join(v.Data.FolderPath, v.Data.FileName)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if _, err := v.Data.DbTx.Exec(string(content)); err != nil {
		return err
	}

	sql := v.databaseUtils.BuildSQLStatement("INSERT INTO migrago (version, name, checksum, applied_at) VALUES (%s, %s, %s, %s)")
	if _, err := v.Data.DbTx.Exec(sql, v.Data.Version, v.Data.FileName, buildChecksum(content), time.Now().UTC()); err != nil {
		return err
	}

	return nil
}

func buildChecksum(content []byte) string {
	s := sha256.Sum256(content)
	return fmt.Sprintf("%x", s)
}
