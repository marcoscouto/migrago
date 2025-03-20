package processor

import (
	"fmt"

	"github.com/marcoscouto/migrago/internal/data"
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
		return fmt.Errorf("invalid migration filename format")
	}
	return v.Next()
}
