package migrago

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Setup de logger
	logger := log.Default()

	// Ler os arquivos da pasta migrations e listar os nomes
	files, err := os.ReadDir("./migrations")
	if err != nil {
		logger.Fatal(err)
	}

	// Conectar ao banco de dados
	psqlInfo := "host=localhost port=5432 user=root password=pass dbname=migrago sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal(err)
	}

	// Criar transação
	transaction, err := db.Begin()
	if err != nil {
		logger.Fatal(err)
	}
	defer transaction.Rollback()

	// Buscar as migrations já executadas
	result, err := transaction.Query("SELECT version, name, checksum, applied_at FROM migrago ORDER BY version DESC")
	if err != nil {
		logger.Fatal(err)
	}

	// Criar um map com as migrations já executadas
	migrationsExecuted := make(map[uint64]Migration)
	var expectedMigration uint64
	for result.Next() {
		var migration Migration
		err = result.Scan(&migration.Version, &migration.Name, &migration.Checksum, &migration.AppliedAt)
		if err != nil {
			logger.Fatal(err)
		}

		if migration.Version > expectedMigration {
			expectedMigration = migration.Version
		}

		migrationsExecuted[migration.Version] = migration
	}

	if err := result.Close(); err != nil {
		logger.Fatal(err)
	}

	// Definir o pattern dos arquivos de migrations
	pattern := regexp.MustCompile(`^(\d+)_([a-zA-Z0-9_-]+)\.sql$`)
	versions := make(map[uint64]bool)

	for _, f := range files {
		// Verificar se todos os arquivos tem o pattern correto (V0_nome.sql)
		match := pattern.Match([]byte(f.Name()))
		if !match {
			err := fmt.Errorf("invalid migration filename format: %s", f.Name())
			logger.Fatal(err)
		}

		// Verificar se todos os arquivos e se não existe repetição
		matches := pattern.FindStringSubmatch(f.Name())
		version, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			logger.Fatal(err)
		}

		if _, ok := versions[version]; ok {
			err := fmt.Errorf("duplicated migration version %d", version)
			logger.Fatal(err)
		}
		versions[version] = true

		// Verificar se a migration já foi executada
		if _, ok := migrationsExecuted[version]; ok {
			logger.Printf("the migration %s already executed\n", f.Name())
			continue
		}

		// Verificar se a migration segue a ordem correta
		if version != expectedMigration+1 {
			err := fmt.Errorf("the migration %s is out of order", f.Name())
			logger.Fatal(err)
		}
		expectedMigration++

		// Executar as migrations
		path := filepath.Join("migrations", f.Name())
		byteContent, err := os.ReadFile(path)
		if err != nil {
			logger.Fatal(err)
		}

		content := string(byteContent)
		if _, err := transaction.Exec(content); err != nil {
			logger.Fatal(err)
		}

		// Criar checksum da migration
		s := sha256.Sum256(byteContent)
		checksum := fmt.Sprintf("%x", s)

		// Salvar o resultado da execução das migrations
		if _, err := transaction.Exec("INSERT INTO migrago (version, name, checksum, applied_at) VALUES ($1, $2, $3, $4)", version, f.Name(), checksum, time.Now().UTC()); err != nil {
			logger.Fatal(err)
		}

		logger.Printf("the migration %s executed successfully\n", f.Name())
	}

	// Commitar as migrations
	if err := transaction.Commit(); err != nil {
		logger.Fatal(err)
	}

	logger.Printf("all migrations executed successfully")
}

type Migration struct {
	Version   uint64
	Name      string
	Checksum  string
	AppliedAt time.Time
}
