package processor

import (
	"fmt"

	"github.com/marcoscouto/migrago/internal/data"
)

type ValidateOrder struct {
	BaseProcessor
}

func NewValidateOrder(data data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &ValidateOrder{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
	}
}

func (v *ValidateOrder) Execute() error {
	v.Data.LastMigration++
	if v.Data.Version != v.Data.LastMigration {
		return fmt.Errorf("the migration is out of order")
	}
	return v.Next()
}
