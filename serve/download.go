package serve

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// DownloadRules downloads the latest mxlint-rules from GitHub
func DownloadRules(rulesDirectory string, log *logrus.Logger) error {
	log.Infof("Rules directory %s not found. Downloading latest mxlint-rules from GitHub...", rulesDirectory)

	// Create the rules directory
	if err := os.MkdirAll(rulesDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", "https://github.com/mxlint/mxlint-rules.git", "--depth", "1", "temp-mxlint-rules")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone mxlint-rules repository: %w", err)
	}

	// Copy the rules directory from the cloned repository to the target directory
	sourceRulesDir := filepath.Join("temp-mxlint-rules", "rules")

	// Copy all rule files from the source to the destination
	err := filepath.Walk(sourceRulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourceRulesDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(rulesDirectory, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, 0755)
		} else {
			// Copy file
			return copyFile(path, destPath)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to copy rules: %w", err)
	}

	// Clean up the temporary directory
	os.RemoveAll("temp-mxlint-rules")

	log.Infof("Successfully downloaded rules to %s", rulesDirectory)
	return nil
}

// copyFile copies a file from src to dst (unexported helper function)
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}
