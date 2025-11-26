package lint

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const cacheVersion = "v1"

// CacheKey represents the unique identifier for a cached result
type CacheKey struct {
	RuleHash  string `json:"rule_hash"`
	InputHash string `json:"input_hash"`
}

// CachedTestcase represents a cached testcase result
type CachedTestcase struct {
	Version  string    `json:"version"`
	CacheKey CacheKey  `json:"cache_key"`
	Testcase *Testcase `json:"testcase"`
}

// getCacheDir returns the cache directory path
func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(homeDir, ".cache", "mxlint")
	return cacheDir, nil
}

// getCachePath returns the full path to a cache file for a given key
func getCachePath(cacheKey CacheKey) (string, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}

	// Create a unique filename from the combined hashes
	combinedHash := fmt.Sprintf("%s-%s", cacheKey.RuleHash, cacheKey.InputHash)
	filename := fmt.Sprintf("%s.json", combinedHash)
	return filepath.Join(cacheDir, filename), nil
}

// computeFileHash computes SHA256 hash of a file's contents
func computeFileHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash), nil
}

// createCacheKey creates a cache key from rule and input file paths
func createCacheKey(rulePath string, inputFilePath string) (*CacheKey, error) {
	ruleHash, err := computeFileHash(rulePath)
	if err != nil {
		return nil, err
	}

	inputHash, err := computeFileHash(inputFilePath)
	if err != nil {
		return nil, err
	}

	return &CacheKey{
		RuleHash:  ruleHash,
		InputHash: inputHash,
	}, nil
}

// loadCachedTestcase loads a testcase from cache if it exists
func loadCachedTestcase(cacheKey CacheKey) (*Testcase, bool) {
	cachePath, err := getCachePath(cacheKey)
	if err != nil {
		log.Debugf("Error getting cache path: %v", err)
		return nil, false
	}

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		log.Debugf("Cache miss: %s", cachePath)
		return nil, false
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		log.Debugf("Error reading cache file: %v", err)
		return nil, false
	}

	// Unmarshal cached data
	var cached CachedTestcase
	if err := json.Unmarshal(data, &cached); err != nil {
		log.Debugf("Error unmarshaling cache: %v", err)
		return nil, false
	}

	// Verify cache version
	if cached.Version != cacheVersion {
		log.Debugf("Cache version mismatch: expected %s, got %s", cacheVersion, cached.Version)
		return nil, false
	}

	// Verify cache key matches
	if cached.CacheKey.RuleHash != cacheKey.RuleHash || cached.CacheKey.InputHash != cacheKey.InputHash {
		log.Debugf("Cache key mismatch")
		return nil, false
	}

	log.Debugf("Cache hit: %s", cachePath)
	return cached.Testcase, true
}

// saveCachedTestcase saves a testcase to cache
func saveCachedTestcase(cacheKey CacheKey, testcase *Testcase) error {
	cachePath, err := getCachePath(cacheKey)
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	// Create cached data structure
	cached := CachedTestcase{
		Version:  cacheVersion,
		CacheKey: cacheKey,
		Testcase: testcase,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}

	// Write to cache file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return err
	}

	log.Debugf("Cached result: %s", cachePath)
	return nil
}

// ClearCache removes all cached files
func ClearCache() error {
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		log.Infof("Cache directory does not exist: %s", cacheDir)
		return nil
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("error removing cache directory: %w", err)
	}

	log.Infof("Cache cleared: %s", cacheDir)
	return nil
}

// GetCacheStats returns statistics about the cache
func GetCacheStats() (int, int64, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return 0, 0, err
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return 0, 0, nil
	}

	var fileCount int
	var totalSize int64

	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	return fileCount, totalSize, err
}

