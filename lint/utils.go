package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New() // Initialize with a default logger

// SetLogger allows the main application to set the logger, including its configuration.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

func expandPaths(pattern string, workingDirectory string) ([]string, error) {
	// backwards compatible with old filepath.glob(...)
	if !strings.HasPrefix(pattern, ".*") {
		oldPattern := pattern
		pattern = strings.ReplaceAll(pattern, "$", "\\$")
		pattern = strings.ReplaceAll(pattern, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "**", ".*")
		log.Infof("Expanded old pattern: %v -> %v", oldPattern, pattern)
	}
	// First get all files recursively under working directory
	var matches []string
	err := filepath.Walk(workingDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Get relative path from working directory
		relPath, err := filepath.Rel(workingDirectory, path)
		if err != nil {
			return err
		}
		// Check if path matches pattern
		matched, err := regexp.MatchString(pattern, relPath)
		if err != nil {
			log.Errorf("Error matching path %v against pattern %v: %v", relPath, pattern, err)
			return err
		}
		if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		log.Warnf("No matches found for pattern %v ", pattern)
	}
	return matches, nil
}
