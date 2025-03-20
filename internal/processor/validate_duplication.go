package processor

import (
	"fmt"

	"github.com/marcoscouto/migrago/internal/data"
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
		return fmt.Errorf("duplicated migration file")
	}
	v.Data.UniqueMigrations[v.Data.Version] = true
	return v.Next()
}
