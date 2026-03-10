package lint

import (
	"archive/zip"
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

func TestSyncSingleRuleset_UnsupportedSource(t *testing.T) {
	err := syncSingleRuleset("ftp://example.com/rules.zip", t.TempDir(), t.TempDir())
	if err == nil {
		t.Fatal("expected unsupported ruleset source error")
	}
}

func TestUnzipToDir_RejectsZipSlipPath(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "rules.zip")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	zipWriter := zip.NewWriter(zipFile)
	entry, err := zipWriter.Create("../evil.txt")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := entry.Write([]byte("evil")); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}

	err = unzipToDir(zipPath, filepath.Join(tempDir, "dest"))
	if err == nil {
		t.Fatal("expected zip slip path error")
	}
}

func TestSyncSingleRuleset_FileSchemeRelativePath(t *testing.T) {
	projectDir := t.TempDir()
	sourceDir := filepath.Join(projectDir, "source-rules")
	targetDir := filepath.Join(projectDir, "target-rules")
	if err := os.MkdirAll(filepath.Join(sourceDir, "mod"), 0755); err != nil {
		t.Fatalf("failed to create source rules dir: %v", err)
	}
	srcFile := filepath.Join(sourceDir, "mod", "rule.rego")
	if err := os.WriteFile(srcFile, []byte("# METADATA\n# title: test\n"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := syncSingleRuleset("file://source-rules", targetDir, projectDir); err != nil {
		t.Fatalf("syncSingleRuleset returned error: %v", err)
	}

	dstFile := filepath.Join(targetDir, "mod", "rule.rego")
	if _, err := os.Stat(dstFile); err != nil {
		t.Fatalf("expected copied file at %s: %v", dstFile, err)
	}
}

func TestCopyRulesFromPath_SourceMustBeDirectory(t *testing.T) {
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "single.rego")
	if err := os.WriteFile(sourceFile, []byte("package test"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	err := copyRulesFromPath(sourceFile, filepath.Join(tempDir, "dest"))
	if err == nil {
		t.Fatal("expected error for non-directory source path")
	}
}
