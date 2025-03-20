package processor

import (
	"strconv"

	"github.com/marcoscouto/migrago/internal/data"
)

type GetVersion struct {
	BaseProcessor
}

func NewGetVersion(data data.MigrationProcessorData, next MigrationProcessor) MigrationProcessor {
	return &GetVersion{
		BaseProcessor: BaseProcessor{
			Data:          data,
			NextProcessor: next,
		},
	}
}

func (v *GetVersion) Execute() error {
	matches := v.Data.Regex.FindStringSubmatch(v.Data.FileName)
	version, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return err
	}
	v.Data.Version = version
	return v.Next()
}
