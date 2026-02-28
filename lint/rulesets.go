package lint

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func SyncRulesets(rulesets []string, rulesPath string, projectDir string) error {
	if len(rulesets) == 0 {
		return nil
	}
	if strings.TrimSpace(rulesPath) == "" {
		return fmt.Errorf("rules path is required when rulesets are configured")
	}

	if err := os.RemoveAll(rulesPath); err != nil {
		return fmt.Errorf("failed to reset rules path %s: %w", rulesPath, err)
	}
	if err := os.MkdirAll(rulesPath, 0755); err != nil {
		return fmt.Errorf("failed to create rules path %s: %w", rulesPath, err)
	}

	for _, ruleset := range rulesets {
		if err := syncSingleRuleset(ruleset, rulesPath, projectDir); err != nil {
			return err
		}
	}
	return nil
}

func syncSingleRuleset(ruleset string, targetPath string, projectDir string) error {
	switch {
	case strings.HasPrefix(ruleset, "file://"):
		localPath := strings.TrimPrefix(ruleset, "file://")
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(projectDir, localPath)
		}
		return copyRulesFromPath(localPath, targetPath)
	case strings.HasPrefix(ruleset, "git://"):
		repoPath := strings.TrimPrefix(ruleset, "git://")
		repoURL := "https://" + strings.TrimPrefix(repoPath, "/")
		return syncGitRuleset(repoURL, targetPath)
	case strings.HasPrefix(ruleset, "https://"), strings.HasPrefix(ruleset, "http://"):
		return syncHTTPRuleset(ruleset, targetPath)
	default:
		return fmt.Errorf("unsupported ruleset source: %s", ruleset)
	}
}

func syncGitRuleset(repoURL string, targetPath string) error {
	tempDir, err := os.MkdirTemp("", "mxlint-ruleset-git-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tempDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone git ruleset %s: %w (%s)", repoURL, err, strings.TrimSpace(string(output)))
	}

	return copyRulesFromPath(selectRulesRoot(tempDir), targetPath)
}

func syncHTTPRuleset(url string, targetPath string) error {
	if strings.HasSuffix(strings.ToLower(url), ".zip") {
		return syncZipRuleset(url, targetPath)
	}
	return fmt.Errorf("unsupported HTTP ruleset source (expected .zip): %s", url)
}

func syncZipRuleset(url string, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download ruleset zip %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to download ruleset zip %s: status %d", url, resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "mxlint-ruleset-*.zip")
	if err != nil {
		return err
	}
	tempZipPath := tempFile.Name()
	defer os.Remove(tempZipPath)
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save ruleset zip: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "mxlint-ruleset-unzip-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	if err := unzipToDir(tempZipPath, tempDir); err != nil {
		return err
	}
	return copyRulesFromPath(selectRulesRoot(tempDir), targetPath)
}

func unzipToDir(zipPath string, destination string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip %s: %w", zipPath, err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetFilePath := filepath.Join(destination, file.Name)
		cleanDestination := filepath.Clean(destination) + string(os.PathSeparator)
		cleanTarget := filepath.Clean(targetFilePath)
		if !strings.HasPrefix(cleanTarget, cleanDestination) {
			return fmt.Errorf("invalid zip path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.Create(cleanTarget)
		if err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return err
		}
		src.Close()
		dst.Close()
	}
	return nil
}

func selectRulesRoot(root string) string {
	candidates := []string{
		filepath.Join(root, "rules"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return root
}

func copyRulesFromPath(sourcePath string, targetPath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("ruleset path %s not found: %w", sourcePath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("ruleset path must be a directory: %s", sourcePath)
	}

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		destPath := filepath.Join(targetPath, relPath)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		return copyFile(path, destPath)
	})
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
