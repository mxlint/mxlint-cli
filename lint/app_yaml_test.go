package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOriginalPathMapMissingFile(t *testing.T) {
	pathMap := loadOriginalPathMap(t.TempDir())
	if len(pathMap) != 0 {
		t.Fatalf("expected empty map for missing app.yaml, got %v", pathMap)
	}
}

func TestLoadOriginalPathMapAndResolve(t *testing.T) {
	tmpDir := t.TempDir()
	appYaml := `
content: []
files:
  - path: Module2/Folder_TRUNCATED/Constant.yaml
    originalPath: Module2/Folder_very_long_name/Constant.yaml
  - path: Metadata.yaml
    originalPath: Metadata.yaml
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.yaml"), []byte(appYaml), 0644); err != nil {
		t.Fatalf("write app.yaml: %v", err)
	}

	pathMap := loadOriginalPathMap(tmpDir)
	if got := resolveOriginalPath("Module2/Folder_TRUNCATED/Constant.yaml", pathMap); got != "Module2/Folder_very_long_name/Constant.yaml" {
		t.Fatalf("mapped path = %q", got)
	}
	if got := resolveOriginalPath("Metadata.yaml", pathMap); got != "Metadata.yaml" {
		t.Fatalf("identity mapped path = %q", got)
	}
	if got := resolveOriginalPath("Unmapped/Doc.yaml", pathMap); got != "Unmapped/Doc.yaml" {
		t.Fatalf("unmapped fallback = %q", got)
	}
}

func TestResolveOriginalPathNilMap(t *testing.T) {
	if got := resolveOriginalPath("Module2/Doc.yaml", nil); got != "Module2/Doc.yaml" {
		t.Fatalf("nil map fallback = %q", got)
	}
}

func TestFormatTestcaseNameWithOriginalPath(t *testing.T) {
	modelSource := t.TempDir()
	inputFile := filepath.Join(modelSource, "Module2", "Doc.yaml")
	name := formatTestcaseName(inputFile, modelSource)
	pathMap := map[string]string{
		"Module2/Doc.yaml": "Module2/Original Doc.yaml",
	}
	if got := resolveOriginalPath(name, pathMap); got != "Module2/Original Doc.yaml" {
		t.Fatalf("originalPath = %q, name = %q", got, name)
	}
}
