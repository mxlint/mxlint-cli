package main

import (
	"flag"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/cinaq/mendix-model-exporter/mpr"
	"github.com/sirupsen/logrus"
)

func main() {

	inputDirectory := flag.String("input", ".", "Directory path")
	outputDirectory := flag.String("output", "modelsource", "Output directory path")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	log := logrus.New()
	if *debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	mpr.SetLogger(log)

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
