package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAllowedRoot(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "doc.yaml")
	if err := os.WriteFile(filePath, []byte("Name: test"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	if got := resolveAllowedRoot(tempDir); got != tempDir {
		t.Fatalf("expected directory root %q, got %q", tempDir, got)
	}
	if got := resolveAllowedRoot(filePath); got != tempDir {
		t.Fatalf("expected file parent root %q, got %q", tempDir, got)
	}
	nonExisting := filepath.Join(tempDir, "missing")
	if got := resolveAllowedRoot(nonExisting); got != nonExisting {
		t.Fatalf("expected passthrough root %q, got %q", nonExisting, got)
	}
}

func TestResolvePath(t *testing.T) {
	tempDir := t.TempDir()
	allowed := filepath.Join(tempDir, "model")
	if err := os.MkdirAll(filepath.Join(allowed, "nested"), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	t.Run("relative path within root", func(t *testing.T) {
		got, err := resolvePath("nested/file.txt", allowed, allowed)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := filepath.Join(allowed, "nested", "file.txt")
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})

	t.Run("absolute path outside root rejected", func(t *testing.T) {
		_, err := resolvePath(filepath.Join(tempDir, "outside.txt"), allowed, allowed)
		if err == nil {
			t.Fatal("expected path outside root error")
		}
	})

	t.Run("empty allowed root falls back to working directory", func(t *testing.T) {
		got, err := resolvePath("nested/again.txt", allowed, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := filepath.Join(allowed, "nested", "again.txt")
		if got != expected {
			t.Fatalf("expected %q, got %q", expected, got)
		}
	})
}
