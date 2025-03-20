package data

import (
	"database/sql"
	"regexp"
)

type MigrationProcessorData struct {
	Regex              *regexp.Regexp
	Version            uint64
	UniqueMigrations   map[uint64]bool
	ExecutedMigrations map[uint64]Migration
	FileName           string
	FolderPath         string
	LastMigration      uint64
	DbTx               *sql.Tx
}
