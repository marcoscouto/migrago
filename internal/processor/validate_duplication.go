package processor

import (
	"github.com/marcoscouto/migrago/internal/data"
	"github.com/marcoscouto/migrago/internal/errors"
)

type validateDuplication struct {
	BaseProcessor
}

func NewValidateDuplication(data *data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &validateDuplication{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
	}
}

func (v *validateDuplication) Execute() error {
	if v.Data.UniqueMigrations[v.Data.Version] {
		return errors.ErrDuplicatedFile
	}
	v.Data.UniqueMigrations[v.Data.Version] = true
	return v.Next()
}
