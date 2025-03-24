package processor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marcoscouto/goql"
	"github.com/marcoscouto/migrago/internal/data"
)

type ExecuteMigration struct {
	BaseProcessor
	goql goql.GoQL
}

func NewExecuteMigration(data *data.MigrationProcessorData, next MigrationProcessor, goql goql.GoQL) MigrationProcessor {
	return &ExecuteMigration{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
		goql: goql,
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

	sql, args := v.goql.BuildSQLStatement("INSERT INTO migrago (version, name, checksum, applied_at) VALUES (%s, %s, %s, %s)", v.Data.Version, v.Data.FileName, buildChecksum(content), time.Now().UTC())
	if _, err := v.Data.DbTx.Exec(sql, args...); err != nil {
		return err
	}

	return nil
}

func buildChecksum(content []byte) string {
	s := sha256.Sum256(content)
	return fmt.Sprintf("%x", s)
}
