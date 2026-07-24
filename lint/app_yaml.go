package lint

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type appYamlFileEntry struct {
	Path         string `yaml:"path"`
	OriginalPath string `yaml:"originalPath"`
}

type appYamlStructure struct {
	Files []appYamlFileEntry `yaml:"files"`
}

// loadOriginalPathMap reads modelsource/app.yaml and returns disk path → original path.
// Missing or unreadable app.yaml yields an empty map (callers fall back to identity).
func loadOriginalPathMap(modelSourcePath string) map[string]string {
	appYamlPath := filepath.Join(modelSourcePath, "app.yaml")
	data, err := os.ReadFile(appYamlPath)
	if err != nil {
		return map[string]string{}
	}

	var structure appYamlStructure
	if err := yaml.Unmarshal(data, &structure); err != nil {
		log.Debugf("Could not parse %s for originalPath mapping: %v", appYamlPath, err)
		return map[string]string{}
	}

	pathMap := make(map[string]string, len(structure.Files))
	for _, entry := range structure.Files {
		path := filepath.ToSlash(strings.TrimSpace(entry.Path))
		if path == "" {
			continue
		}
		original := filepath.ToSlash(strings.TrimSpace(entry.OriginalPath))
		if original == "" {
			original = path
		}
		pathMap[path] = original
	}
	return pathMap
}

// resolveOriginalPath returns the mapped original path, or name when unmapped.
func resolveOriginalPath(name string, pathMap map[string]string) string {
	name = filepath.ToSlash(name)
	if name == "" {
		return name
	}
	if pathMap != nil {
		if original, ok := pathMap[name]; ok && original != "" {
			return original
		}
	}
	return name
}
