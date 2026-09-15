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

func TestEnsureGitRepositoryInitializesModelsourceRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "modelsource")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("failed to create model directory: %v", err)
	}

	created, err := EnsureGitRepository(modelDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected a new git repository to be created")
	}

	isRoot, err := isGitRoot(modelDir)
	if err != nil {
		t.Fatalf("failed to check git root: %v", err)
	}
	if !isRoot {
		t.Fatal("expected modelsource to be a git root")
	}

	createdAgain, err := EnsureGitRepository(modelDir)
	if err != nil {
		t.Fatalf("unexpected error on second ensure: %v", err)
	}
	if createdAgain {
		t.Fatal("expected ensure to be idempotent")
	}
}

func TestEnsureGitRepositoryCreatesMissingDirectory(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "modelsource")

	created, err := EnsureGitRepository(modelDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected a new git repository to be created")
	}

	info, err := os.Stat(modelDir)
	if err != nil {
		t.Fatalf("expected modelsource directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected modelsource path to be a directory")
	}

	isRoot, err := isGitRoot(modelDir)
	if err != nil {
		t.Fatalf("failed to check git root: %v", err)
	}
	if !isRoot {
		t.Fatal("expected modelsource to be a git root")
	}
}

func TestEnsureGitRepositoryCreatesNestedRootInsideParentRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	tempDir := t.TempDir()
	runGit(t, tempDir, "init")

	modelDir := filepath.Join(tempDir, "modelsource")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("failed to create model directory: %v", err)
	}

	created, err := EnsureGitRepository(modelDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected nested modelsource git repository to be created")
	}

	topLevel, err := gitTopLevel(modelDir)
	if err != nil {
		t.Fatalf("failed to resolve git toplevel: %v", err)
	}
	if cleanPath(topLevel) != cleanPath(modelDir) {
		t.Fatalf("expected git toplevel %q, got %q", cleanPath(modelDir), cleanPath(topLevel))
	}
}

func TestPersistGitRepositoryCommitsModelsourceState(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "modelsource")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("failed to create model directory: %v", err)
	}

	docPath := filepath.Join(modelDir, "Security$ProjectSecurity.yaml")
	if err := os.WriteFile(docPath, []byte("name: initial\n"), 0644); err != nil {
		t.Fatalf("failed to write model document: %v", err)
	}

	committed, err := PersistGitRepository(modelDir, "snapshot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !committed {
		t.Fatal("expected modelsource changes to be committed")
	}

	committedAgain, err := PersistGitRepository(modelDir, "snapshot")
	if err != nil {
		t.Fatalf("unexpected error on second persist: %v", err)
	}
	if committedAgain {
		t.Fatal("expected nothing to commit on second persist")
	}

	if err := os.WriteFile(docPath, []byte("name: changed\n"), 0644); err != nil {
		t.Fatalf("failed to update model document: %v", err)
	}

	changedFiles, err := GitUnstagedChangedFiles(modelDir)
	if err != nil {
		t.Fatalf("unexpected error listing changes: %v", err)
	}
	if len(changedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d (%v)", len(changedFiles), changedFiles)
	}
	if changedFiles[0] != cleanPath(docPath) {
		t.Fatalf("expected %q, got %q", cleanPath(docPath), changedFiles[0])
	}
}

func TestGitUnstagedChangedFilesDetectsUntrackedModelDocument(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "modelsource")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("failed to create model directory: %v", err)
	}

	if _, err := EnsureGitRepository(modelDir); err != nil {
		t.Fatalf("failed to ensure git repository: %v", err)
	}

	docPath := filepath.Join(modelDir, "Security$ProjectSecurity.yaml")
	if err := os.WriteFile(docPath, []byte("name: initial\n"), 0644); err != nil {
		t.Fatalf("failed to write model document: %v", err)
	}

	changedFiles, err := GitUnstagedChangedFiles(modelDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changedFiles) != 1 {
		t.Fatalf("expected 1 untracked file, got %d (%v)", len(changedFiles), changedFiles)
	}
	if changedFiles[0] != cleanPath(docPath) {
		t.Fatalf("expected %q, got %q", cleanPath(docPath), changedFiles[0])
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
