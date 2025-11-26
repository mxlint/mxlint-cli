package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCaching(t *testing.T) {
	// Create a temporary cache directory for testing
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Test computeFileHash
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash1, err := computeFileHash(testFile)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	// Hash should be consistent
	hash2, err := computeFileHash(testFile)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Hash mismatch: %s != %s", hash1, hash2)
	}

	// Different content should produce different hash
	if err := os.WriteFile(testFile, []byte("different content"), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	hash3, err := computeFileHash(testFile)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}

	if hash1 == hash3 {
		t.Errorf("Hash should be different for different content")
	}
}

func TestCacheKeyCreation(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create test files
	ruleFile := filepath.Join(tempDir, "rule.rego")
	inputFile := filepath.Join(tempDir, "input.yaml")

	if err := os.WriteFile(ruleFile, []byte("package test"), 0644); err != nil {
		t.Fatalf("Failed to create rule file: %v", err)
	}

	if err := os.WriteFile(inputFile, []byte("test: data"), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Create cache key
	cacheKey, err := createCacheKey(ruleFile, inputFile)
	if err != nil {
		t.Fatalf("Failed to create cache key: %v", err)
	}

	if cacheKey.RuleHash == "" {
		t.Error("RuleHash should not be empty")
	}

	if cacheKey.InputHash == "" {
		t.Error("InputHash should not be empty")
	}
}

func TestCacheLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create a cache key
	cacheKey := CacheKey{
		RuleHash:  "abc123",
		InputHash: "def456",
	}

	// Create a testcase
	testcase := &Testcase{
		Name: "test-case",
		Time: 1.5,
		Failure: &Failure{
			Message: "test failure",
			Type:    "error",
		},
	}

	// Save to cache
	if err := saveCachedTestcase(cacheKey, testcase); err != nil {
		t.Fatalf("Failed to save to cache: %v", err)
	}

	// Load from cache
	loadedTestcase, found := loadCachedTestcase(cacheKey)
	if !found {
		t.Fatal("Testcase should be found in cache")
	}

	if loadedTestcase.Name != testcase.Name {
		t.Errorf("Name mismatch: expected %s, got %s", testcase.Name, loadedTestcase.Name)
	}

	if loadedTestcase.Time != testcase.Time {
		t.Errorf("Time mismatch: expected %f, got %f", testcase.Time, loadedTestcase.Time)
	}

	if loadedTestcase.Failure == nil {
		t.Fatal("Failure should not be nil")
	}

	if loadedTestcase.Failure.Message != testcase.Failure.Message {
		t.Errorf("Failure message mismatch: expected %s, got %s", testcase.Failure.Message, loadedTestcase.Failure.Message)
	}
}

func TestCacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create a cache key that doesn't exist
	cacheKey := CacheKey{
		RuleHash:  "nonexistent",
		InputHash: "nothere",
	}

	// Should not find anything
	_, found := loadCachedTestcase(cacheKey)
	if found {
		t.Error("Should not find non-existent cache entry")
	}
}

func TestClearCache(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Save something to cache
	cacheKey := CacheKey{
		RuleHash:  "test123",
		InputHash: "input456",
	}

	testcase := &Testcase{
		Name: "test",
		Time: 1.0,
	}

	if err := saveCachedTestcase(cacheKey, testcase); err != nil {
		t.Fatalf("Failed to save to cache: %v", err)
	}

	// Verify it exists
	_, found := loadCachedTestcase(cacheKey)
	if !found {
		t.Fatal("Cache entry should exist")
	}

	// Clear cache
	if err := ClearCache(); err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify it's gone
	_, found = loadCachedTestcase(cacheKey)
	if found {
		t.Error("Cache entry should be cleared")
	}
}

func TestGetCacheStats(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Initially should have no cache
	count, size, err := GetCacheStats()
	if err != nil {
		t.Fatalf("Failed to get cache stats: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 files, got %d", count)
	}

	if size != 0 {
		t.Errorf("Expected 0 size, got %d", size)
	}

	// Save some cache entries
	for i := 0; i < 5; i++ {
		cacheKey := CacheKey{
			RuleHash:  fmt.Sprintf("rule%d", i),
			InputHash: fmt.Sprintf("input%d", i),
		}

		testcase := &Testcase{
			Name: "test",
			Time: float64(i),
		}

		if err := saveCachedTestcase(cacheKey, testcase); err != nil {
			t.Fatalf("Failed to save to cache: %v", err)
		}
	}

	// Check stats
	count, size, err = GetCacheStats()
	if err != nil {
		t.Fatalf("Failed to get cache stats: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected 5 files, got %d", count)
	}

	if size == 0 {
		t.Error("Size should be greater than 0")
	}
}

