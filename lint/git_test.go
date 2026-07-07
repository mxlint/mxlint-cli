package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitUnstagedChangedFilesRequiresRepository(t *testing.T) {
	tempDir := t.TempDir()
	_, err := GitUnstagedChangedFiles(tempDir)
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
	if err != ErrNotGitRepository {
		t.Fatalf("expected ErrNotGitRepository, got %v", err)
	}
}

func TestGitUnstagedChangedFilesDetectsUnstagedModelDocument(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	tempDir := t.TempDir()
	runGit(t, tempDir, "init")
	runGit(t, tempDir, "config", "user.email", "mxlint@test.local")
	runGit(t, tempDir, "config", "user.name", "mxlint test")

	modelDir := filepath.Join(tempDir, "modelsource")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("failed to create model directory: %v", err)
	}

	docPath := filepath.Join(modelDir, "Security$ProjectSecurity.yaml")
	if err := os.WriteFile(docPath, []byte("name: initial\n"), 0644); err != nil {
		t.Fatalf("failed to write model document: %v", err)
	}
	runGit(t, tempDir, "add", "modelsource/Security$ProjectSecurity.yaml")
	runGit(t, tempDir, "commit", "-m", "initial")

	if err := os.WriteFile(docPath, []byte("name: changed\n"), 0644); err != nil {
		t.Fatalf("failed to update model document: %v", err)
	}

	changedFiles, err := GitUnstagedChangedFiles(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d (%v)", len(changedFiles), changedFiles)
	}
	if changedFiles[0] != cleanPath(docPath) {
		t.Fatalf("expected %q, got %q", cleanPath(docPath), changedFiles[0])
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, string(output))
	}
}
