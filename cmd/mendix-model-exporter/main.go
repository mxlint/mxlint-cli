package main

import (
	"flag"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/cinaq/mendix-model-exporter/mpr"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	inputDirectory := flag.String("input", ".", "Directory path")
	outputDirectory := flag.String("output", "modelsource", "Output directory path")
	flag.Parse()

	err := filepath.Walk(*inputDirectory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".mpr") {
			mpr.Export(path, *outputDirectory)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking through directory: %v", err)
	}
}
