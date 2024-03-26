package mpr

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New() // Initialize with a default logger

// SetLogger allows the main application to set the logger, including its configuration.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
