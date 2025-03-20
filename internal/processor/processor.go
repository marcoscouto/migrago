package processor

import (
	"github.com/marcoscouto/migrago/internal/data"
)

type MigrationProcessor interface {
	Execute() error
}

type BaseProcessor struct {
	Data          data.MigrationProcessorData
	NextProcessor MigrationProcessor
}

func (p BaseProcessor) Next() error {
	if p.NextProcessor == nil {
		return nil
	}
	return p.NextProcessor.Execute()
}
