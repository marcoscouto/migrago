package processor

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/marcoscouto/migrago/internal/data"
)

type ExecuteMigration struct {
	BaseProcessor
}

func NewExecuteMigration(data data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &ExecuteMigration{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
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

	if _, err := v.Data.DbTx.Exec("INSERT INTO migrago (version, name, checksum, applied_at) VALUES ($1, $2, $3, $4)", v.Data.Version, v.Data.FileName, buildChecksum(content), time.Now().UTC()); err != nil {
		return err
	}

	return nil
}

func buildChecksum(content []byte) string {
	s := sha256.Sum256(content)
	return fmt.Sprintf("%x", s)
}
