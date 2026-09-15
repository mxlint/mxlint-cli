package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterFilesUnderDirectory(t *testing.T) {
	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "modelsource")
	nestedDir := filepath.Join(modelDir, "Module2")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	inside := filepath.Join(nestedDir, "DomainModels$DomainModel.yaml")
	outside := filepath.Join(tempDir, "mxlint.yaml")
	for _, path := range []string{inside, outside} {
		if err := os.WriteFile(path, []byte("name: test\n"), 0644); err != nil {
			t.Fatalf("failed to write file %q: %v", path, err)
		}
	}

	filtered, err := FilterFilesUnderDirectory([]string{inside, outside}, modelDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("expected 1 file, got %d (%v)", len(filtered), filtered)
	}
	if filtered[0] != cleanPath(inside) {
		t.Fatalf("expected %q, got %q", cleanPath(inside), filtered[0])
	}
}

func TestFilterInputFiles(t *testing.T) {
	tempDir := t.TempDir()
	changed := filepath.Join(tempDir, "changed.yaml")
	unchanged := filepath.Join(tempDir, "unchanged.yaml")
	for _, path := range []string{changed, unchanged} {
		if err := os.WriteFile(path, []byte("name: test\n"), 0644); err != nil {
			t.Fatalf("failed to write file %q: %v", path, err)
		}
	}

	changedSet := normalizeChangedFilesSet([]string{changed})
	filtered := filterInputFiles([]string{changed, unchanged}, changedSet)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 file, got %d (%v)", len(filtered), filtered)
	}
	if filtered[0] != changed {
		t.Fatalf("expected %q, got %q", changed, filtered[0])
	}
}

func TestFilterInputFilesNilSetReturnsAll(t *testing.T) {
	inputFiles := []string{"a.yaml", "b.yaml"}
	filtered := filterInputFiles(inputFiles, nil)
	if len(filtered) != len(inputFiles) {
		t.Fatalf("expected all files to be returned, got %v", filtered)
	}
}
