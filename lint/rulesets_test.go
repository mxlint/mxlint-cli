package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyRulesFromPath_SameSourceAndTarget_NoOp(t *testing.T) {
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	if err := os.MkdirAll(filepath.Join(rulesDir, "sample"), 0755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}
	rulePath := filepath.Join(rulesDir, "sample", "rule.js")
	original := []byte("const metadata = { title: 'a', description: 'b', custom: { rulenumber: '001_0001' } };")
	if err := os.WriteFile(rulePath, original, 0644); err != nil {
		t.Fatalf("failed to write sample rule: %v", err)
	}

	if err := copyRulesFromPath(rulesDir, rulesDir); err != nil {
		t.Fatalf("copyRulesFromPath should no-op for identical source/target: %v", err)
	}

	got, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatalf("failed to read sample rule after copy: %v", err)
	}
	if string(got) != string(original) {
		t.Fatalf("expected rule content unchanged, got %q", string(got))
	}
}

func TestCopyFile_SameSourceAndDestination_NoOp(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "rule.rego")
	original := []byte("# METADATA\n# title: x\n")
	if err := os.WriteFile(filePath, original, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := copyFile(filePath, filePath); err != nil {
		t.Fatalf("copyFile should no-op for identical source/destination: %v", err)
	}

	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file after copy: %v", err)
	}
	if string(got) != string(original) {
		t.Fatalf("expected file content unchanged, got %q", string(got))
	}
}
