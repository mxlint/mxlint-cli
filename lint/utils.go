package lint

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log = logrus.New() // Initialize with a default logger

// SetLogger allows the main application to set the logger, including its configuration.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

func expandPaths(pattern string, workingDirectory string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(workingDirectory, pattern))
	if err != nil {
		// An error occurred while globbing, return the error.
		return nil, err
	}
	// Return the matches found.
	return matches, nil
}
