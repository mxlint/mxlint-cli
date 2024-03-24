package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: program <directory_path>")
	}
	directoryPath := os.Args[1]

	err := filepath.Walk(directoryPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".db") {
			processDBFile(path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking through directory: %v", err)
	}
}

func processDBFile(dbFilePath string) {
	fmt.Printf("Processing %s\n", dbFilePath)

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT content FROM documents")
	if err != nil {
		log.Fatalf("Error querying documents: %v", err)
	}
	defer rows.Close()

	var counter int
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		var jsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(content), &jsonObj); err != nil {
			log.Fatalf("Error unmarshaling JSON: %v", err)
		}

		yamlData, err := yaml.Marshal(jsonObj)
		if err != nil {
			log.Fatalf("Error marshaling YAML: %v", err)
		}

		fileName := fmt.Sprintf("%s_document_%d.yaml", filepath.Base(dbFilePath), counter)
		if err := os.WriteFile(fileName, yamlData, 0644); err != nil {
			log.Fatalf("Error writing YAML file: %v", err)
		}
		fmt.Printf("Wrote %s\n", fileName)
		counter++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}
}
