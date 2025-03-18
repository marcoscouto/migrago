package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	// Ler os arquivos da pasta migrations e listar os nomes
	files, err := os.ReadDir("./migrations")
	if err != nil {
		panic(err)
	}

	versions := make(map[string]bool)

	for _, f := range files {
		// Verificar se todos os arquivos tem o pattern correto (V0_nome.sql)
		pattern, err := regexp.Compile(`^(\d+)_(\w+)\.sql$`)
		if err != nil {
			panic(err)
		}

		match := pattern.Match([]byte(f.Name()))
		if !match {
			err := fmt.Errorf("the pattern of %s don't match", f.Name())
			panic(err)
		}

		// Verificar se todos os arquivos e se não existe repetição
		matches := pattern.FindStringSubmatch(f.Name())
		version := matches[1]
		if _, ok := versions[version]; ok {
			err := fmt.Errorf("duplicated version %s", version)
			panic(err)
		}
		versions[version] = true

		// Captura o conteúdo do arquivo
		// content, err := os.ReadFile(fmt.Sprint("./migrations/", f.Name()))
		// if err != nil {
		// 	panic(err)
		// }

		// Verificar se a migration já foi executada
		// Verificar se a migration segue a ordem correta
		// Executar as migrations
		// Salvar o resultado da execução das migrations
	}
	
	fmt.Println("executed successfully")
}