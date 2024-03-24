package mpr

import (
	_ "github.com/mattn/go-sqlite3"
)

func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
