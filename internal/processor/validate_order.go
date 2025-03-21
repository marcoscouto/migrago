package processor

import (
	"github.com/marcoscouto/migrago/internal/data"
	"github.com/marcoscouto/migrago/internal/errors"
)

type ValidateOrder struct {
	BaseProcessor
}

func NewValidateOrder(data *data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &ValidateOrder{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
	}
}

func (v *ValidateOrder) Execute() error {
	*v.Data.LastMigration++
	if v.Data.Version != *v.Data.LastMigration {
		return errors.ErrOutOfOrder
	}
	return v.Next()
}
