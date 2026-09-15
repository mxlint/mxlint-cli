package mpr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkipMendixCacheDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	cacheDir := filepath.Join(root, mendixCacheDirName)
	if err := os.MkdirAll(filepath.Join(cacheDir, "extensions-cache", "nested"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "extensions-cache", "nested", "file.dll"), []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var visited []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := skipMendixCacheDir(path, info); err != nil {
			return err
		}
		visited = append(visited, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	for _, path := range visited {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatalf("rel: %v", err)
		}
		if rel == mendixCacheDirName {
			continue
		}
		if len(rel) > len(mendixCacheDirName) && rel[:len(mendixCacheDirName)+1] == mendixCacheDirName+string(os.PathSeparator) {
			t.Fatalf("walk descended into mendix cache: %s", path)
		}
	}
}
