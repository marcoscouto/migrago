package utils

import (
	"strconv"
	"strings"
)

const (
	MySQL      = "mysql"
	PostgreSQL = "postgres"
	SQLite     = "sqlite"
	SQLServer  = "sqlserver"
	Oracle     = "oracle"
)

type DatabaseUtils interface {
	BuildSQLStatement(content string) string
}

type databaseUtils struct {
	databaseDriver string
}

func NewDatabaseUtils(databaseDriver string) DatabaseUtils {
	return &databaseUtils{
		databaseDriver: databaseDriver,
	}
}

func (d *databaseUtils) BuildSQLStatement(content string) string {
	switch d.databaseDriver {
	case MySQL, SQLite:
		return strings.ReplaceAll(content, "%s", "?")
	case SQLServer:
		return strings.ReplaceAll(content, "%s", "@p")
	case Oracle:
		return strings.ReplaceAll(content, "%s", ":param")
	default:
		result := content
		for i := 1; i <= strings.Count(content, "%s"); i++ {
			result = strings.Replace(result, "%s", "$"+strconv.Itoa(i), 1)
		}
		return result
	}
}