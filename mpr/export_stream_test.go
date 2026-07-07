package mpr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEffectiveExportConcurrency(t *testing.T) {
	t.Parallel()

	SetExportConcurrency(0)
	t.Cleanup(func() {
		SetExportConcurrency(0)
	})

	if got := effectiveExportConcurrency(0); got != 1 {
		t.Fatalf("expected 1 for zero documents, got %d", got)
	}

	SetExportConcurrency(8)
	if got := effectiveExportConcurrency(3); got != 3 {
		t.Fatalf("expected concurrency capped to document count 3, got %d", got)
	}
}

func TestExportPlanLoadUsesCache(t *testing.T) {
	plan := &exportPlan{
		unitCache: map[string]cachedUnitContent{
			"unit-1": {
				Contents:     map[string]interface{}{"Name": "Cached"},
				ContentsHash: "abc123",
			},
		},
	}

	contents, err := plan.loadDocument("unit-1")
	if err != nil {
		t.Fatalf("loadDocument() unexpected error: %v", err)
	}
	if contents["Name"] != "Cached" {
		t.Fatalf("expected cached contents, got %#v", contents)
	}
}

func TestOutputFileMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mpr-output-match-*")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := []byte("name: Example\n")
	path := filepath.Join(tmpDir, "example.yaml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	same, err := outputFileMatches(path, content)
	if err != nil {
		t.Fatalf("outputFileMatches() error: %v", err)
	}
	if !same {
		t.Fatal("expected matching output file to be detected")
	}

	changed, err := outputFileMatches(path, []byte("name: Changed\n"))
	if err != nil {
		t.Fatalf("outputFileMatches() error: %v", err)
	}
	if changed {
		t.Fatal("expected different content to not match")
	}
}

func TestWriteFileWithPersistentCacheSkipsUnchangedOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mpr-output-skip-*")
	if err != nil {
		t.Fatalf("mkdir temp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cacheDir := filepath.Join(tmpDir, "cache")
	SetPersistentYAMLCacheDirectory(cacheDir)
	SetPersistentYAMLCacheEnabled(true)
	t.Cleanup(func() {
		SetPersistentYAMLCacheDirectory("")
		SetPersistentYAMLCacheEnabled(true)
	})

	contents := map[string]interface{}{"Name": "Example"}
	hash := "hash-for-skip-test"
	outPath := filepath.Join(tmpDir, "doc.yaml")

	if err := writeFileWithPersistentCache(outPath, contents, hash, false); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	before, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}

	if err := writeFileWithPersistentCache(outPath, contents, hash, false); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	after, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat output after second write: %v", err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("expected unchanged output file to be left untouched")
	}
}
