package processor

import "github.com/marcoscouto/migrago/internal/data"

type VerifyExecuted struct {
	BaseProcessor
}

func NewVerifyExecuted(data data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &VerifyExecuted{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
	}
}

func (v *VerifyExecuted) Execute() error {
	if _, ok := v.Data.ExecutedMigrations[v.Data.Version]; ok {
		return nil
	}
	return v.Next()
}
