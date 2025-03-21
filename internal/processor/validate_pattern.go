package processor

import (
	"github.com/marcoscouto/migrago/internal/data"
	"github.com/marcoscouto/migrago/internal/errors"
)

type ValidatePattern struct {
	BaseProcessor
}

func NewValidatePattern(data *data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &ValidatePattern{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
	}
}

func (v *ValidatePattern) Execute() error {
	match := v.Data.Regex.Match([]byte(v.Data.FileName))
	if !match {
		return errors.ErrInvalidPattern
	}
	return v.Next()
}
