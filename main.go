package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/marcoscouto/migrago/migrago"
)

func main() {
	// Setup de logger
	logger := log.Default()

	// Criar conexão com banco de dados
	psqlInfo := "host=localhost port=5432 user=root password=pass dbname=migrago sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal(err)
	}

	// Executar migrações
	if err := migrago.New(db).ExecuteMigrations("./migrations"); err != nil {
		logger.Fatal(err)
	}

	logger.Printf("all migrations executed successfully")
}
