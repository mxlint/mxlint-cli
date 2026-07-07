package lint

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrNotGitRepository = errors.New("not a git repository")

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

// GitUnstagedChangedFiles returns absolute paths of files with unstaged changes.
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

	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list unstaged changes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	changedFiles := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		changedFiles = append(changedFiles, cleanPath(filepath.Join(gitRoot, line)))
	}
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
