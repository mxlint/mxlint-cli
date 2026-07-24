package lint

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var ErrNotGitRepository = errors.New("not a git repository")

const (
	mxlintGitUserName  = "mxlint"
	mxlintGitUserEmail = "mxlint@localhost"
	defaultPersistMsg  = "mxlint: commit modelsource"
)

// IsGitRepository reports whether dir is inside a git work tree.
func IsGitRepository(dir string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		if isGitCommandNotFound(err) {
			return false, fmt.Errorf("git is not installed or not available in PATH")
		}
		return false, nil
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

// EnsureGitRepository ensures dir exists and is a git repository root.
// When dir is missing it is created. When dir is not already a git root,
// it runs git init and configures a local identity.
// Returns true when a new repository was created.
func EnsureGitRepository(dir string) (bool, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, fmt.Errorf("failed to resolve directory %q: %w", dir, err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to inspect modelsource directory %q: %w", absDir, err)
		}
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return false, fmt.Errorf("failed to create modelsource directory %q: %w", absDir, err)
		}
	} else if !info.IsDir() {
		return false, fmt.Errorf("modelsource path %q is not a directory", absDir)
	}

	isRoot, err := isGitRoot(absDir)
	if err != nil {
		return false, err
	}
	if isRoot {
		return false, nil
	}

	if _, err := exec.LookPath("git"); err != nil {
		return false, fmt.Errorf("git is not installed or not available in PATH")
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = absDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to initialize git repository in %q: %w (%s)", absDir, err, strings.TrimSpace(string(output)))
	}

	if err := configureMxlintGitIdentity(absDir); err != nil {
		return false, err
	}

	return true, nil
}

// PersistGitRepository stages all changes in dir and creates a commit.
// Ensures the directory is a git repository first.
// Returns false when there is nothing to commit.
func PersistGitRepository(dir string, message string) (bool, error) {
	if _, err := EnsureGitRepository(dir); err != nil {
		return false, err
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false, fmt.Errorf("failed to resolve directory %q: %w", dir, err)
	}

	if strings.TrimSpace(message) == "" {
		message = defaultPersistMsg
	}

	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = absDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to stage modelsource changes: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = absDir
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to inspect modelsource git status: %w", err)
	}
	if strings.TrimSpace(string(statusOutput)) == "" {
		return false, nil
	}

	commitCmd := exec.Command("git",
		"-c", "user.name="+mxlintGitUserName,
		"-c", "user.email="+mxlintGitUserEmail,
		"commit", "-m", message,
	)
	commitCmd.Dir = absDir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("failed to commit modelsource changes: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	return true, nil
}

// GitUnstagedChangedFiles returns absolute paths of files with unstaged changes
// and untracked files (excluding ignored paths).
func GitUnstagedChangedFiles(dir string) ([]string, error) {
	isRepo, err := IsGitRepository(dir)
	if err != nil {
		return nil, err
	}
	if !isRepo {
		return nil, ErrNotGitRepository
	}

	gitRoot, err := gitTopLevel(dir)
	if err != nil {
		return nil, err
	}

	changed := make(map[string]struct{})

	diffCmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR")
	diffCmd.Dir = dir
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list unstaged changes: %w", err)
	}
	for _, line := range splitNonEmptyLines(string(diffOutput)) {
		changed[cleanPath(filepath.Join(gitRoot, line))] = struct{}{}
	}

	untrackedCmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedCmd.Dir = dir
	untrackedOutput, err := untrackedCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list untracked files: %w", err)
	}
	for _, line := range splitNonEmptyLines(string(untrackedOutput)) {
		changed[cleanPath(filepath.Join(gitRoot, line))] = struct{}{}
	}

	changedFiles := make([]string, 0, len(changed))
	for path := range changed {
		changedFiles = append(changedFiles, path)
	}
	sort.Strings(changedFiles)
	return changedFiles, nil
}

// FilterFilesUnderDirectory keeps only files located under directory.
func FilterFilesUnderDirectory(files []string, directory string) ([]string, error) {
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory %q: %w", directory, err)
	}
	absDir = cleanPath(absDir)

	filtered := make([]string, 0, len(files))
	for _, file := range files {
		absFile, err := filepath.Abs(file)
		if err != nil {
			continue
		}
		absFile = cleanPath(absFile)
		rel, err := filepath.Rel(absDir, absFile)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		filtered = append(filtered, absFile)
	}
	return filtered, nil
}

func isGitRoot(dir string) (bool, error) {
	gitPath := filepath.Join(dir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect %q: %w", gitPath, err)
	}
	// .git may be a directory (normal repo) or a file (worktree/submodule).
	return info.IsDir() || info.Mode().IsRegular(), nil
}

func configureMxlintGitIdentity(dir string) error {
	for _, args := range [][]string{
		{"config", "user.name", mxlintGitUserName},
		{"config", "user.email", mxlintGitUserEmail},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to configure git %s: %w (%s)", args[1], err, strings.TrimSpace(string(output)))
		}
	}
	return nil
}

func splitNonEmptyLines(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}
	return result
}

func gitTopLevel(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to resolve git root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func isGitCommandNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() == 127
	}
	return false
}

func cleanPath(path string) string {
	cleaned := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		return cleaned
	}
	return resolved
}
