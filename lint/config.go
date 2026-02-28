package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const configFileName = "mxlint.yaml"

type Config struct {
	Rules  ConfigRulesSpec  `yaml:"rules"`
	Lint   ConfigLintSpec   `yaml:"lint"`
	Export ConfigExportSpec `yaml:"export"`
}

type ConfigRulesSpec struct {
	Path     string   `yaml:"path"`
	Rulesets []string `yaml:"rulesets"`
}

type ConfigExportSpec struct {
	Output string `yaml:"output"`
	Input  string `yaml:"input"`
	Mode   string `yaml:"mode"`
	Filter string `yaml:"filter"`
}

type ConfigLintSpec struct {
	Skip map[string][]ConfigSkipRule `yaml:"skip"`
}

type ConfigSkipRule struct {
	Rule   string `yaml:"rule"`
	Reason string `yaml:"reason"`
	Date   string `yaml:"date"`
}

var activeConfig = struct {
	mu     sync.RWMutex
	config *Config
}{
	config: &Config{},
}

var defaultConfig = struct {
	mu      sync.RWMutex
	content []byte
}{}

func SetDefaultConfigYAML(content []byte) {
	defaultConfig.mu.Lock()
	defer defaultConfig.mu.Unlock()
	if len(content) == 0 {
		defaultConfig.content = nil
		return
	}
	defaultConfig.content = append([]byte{}, content...)
}

func getDefaultConfigYAML() []byte {
	defaultConfig.mu.RLock()
	defer defaultConfig.mu.RUnlock()
	if len(defaultConfig.content) == 0 {
		return nil
	}
	return append([]byte{}, defaultConfig.content...)
}

func SetConfig(config *Config) {
	activeConfig.mu.Lock()
	defer activeConfig.mu.Unlock()
	if config == nil {
		activeConfig.config = &Config{}
		return
	}
	activeConfig.config = config
}

func getConfig() *Config {
	activeConfig.mu.RLock()
	defer activeConfig.mu.RUnlock()
	if activeConfig.config == nil {
		return &Config{}
	}
	return activeConfig.config
}

func LoadMergedConfig(projectDir string) (*Config, error) {
	cfg := &Config{}

	systemConfigPath, err := resolveSystemConfigPath()
	if err != nil {
		return nil, err
	}
	projectConfigPath := filepath.Join(projectDir, configFileName)

	defaultCfg, err := loadConfigYAML(getDefaultConfigYAML(), "embedded default config")
	if err != nil {
		return nil, err
	}
	mergeConfig(cfg, defaultCfg)

	systemCfg, err := loadConfigFile(systemConfigPath)
	if err != nil {
		return nil, err
	}
	mergeConfig(cfg, systemCfg)

	projectCfg, err := loadConfigFile(projectConfigPath)
	if err != nil {
		return nil, err
	}
	mergeConfig(cfg, projectCfg)

	return cfg, nil
}

func resolveSystemConfigPath() (string, error) {
	if explicit := strings.TrimSpace(os.Getenv("MXLINT_SYSTEM_CONFIG")); explicit != "" {
		return explicit, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// In restricted environments HOME/USERPROFILE may be unavailable.
		// Treat this as "no system config".
		return "", nil
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, configFileName), nil
	}

	return filepath.Join(homeDir, ".config", configFileName), nil
}

func loadConfigFile(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, nil
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}
	return loadConfigYAML(content, configPath)
}

func loadConfigYAML(content []byte, source string) (*Config, error) {
	if len(content) == 0 {
		return nil, nil
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", source, err)
	}
	return cfg, nil
}

func mergeConfig(base *Config, overlay *Config) {
	if overlay == nil {
		return
	}
	if strings.TrimSpace(overlay.Rules.Path) != "" {
		base.Rules.Path = strings.TrimSpace(overlay.Rules.Path)
	}
	if len(overlay.Rules.Rulesets) > 0 {
		base.Rules.Rulesets = append([]string{}, overlay.Rules.Rulesets...)
	}

	if strings.TrimSpace(overlay.Export.Output) != "" {
		base.Export.Output = strings.TrimSpace(overlay.Export.Output)
	}
	if strings.TrimSpace(overlay.Export.Input) != "" {
		base.Export.Input = strings.TrimSpace(overlay.Export.Input)
	}
	if strings.TrimSpace(overlay.Export.Mode) != "" {
		base.Export.Mode = strings.TrimSpace(overlay.Export.Mode)
	}
	if overlay.Export.Filter != "" {
		base.Export.Filter = strings.TrimSpace(overlay.Export.Filter)
	}

	if len(overlay.Lint.Skip) == 0 {
		return
	}
	if base.Lint.Skip == nil {
		base.Lint.Skip = map[string][]ConfigSkipRule{}
	}
	for documentPath, entries := range overlay.Lint.Skip {
		base.Lint.Skip[normalizeSkipPath(documentPath)] = append([]ConfigSkipRule{}, entries...)
	}
}

func shouldSkipByConfig(inputFilePath string, ruleNumber string, modelSourcePath string) (bool, string) {
	cfg := getConfig()
	if cfg == nil || len(cfg.Lint.Skip) == 0 {
		return false, ""
	}

	for _, candidate := range buildSkipPathCandidates(inputFilePath, modelSourcePath) {
		entries, ok := cfg.Lint.Skip[candidate]
		if !ok {
			continue
		}

		for _, entry := range entries {
			if entry.Rule == "" || entry.Rule == "*" || entry.Rule == ruleNumber {
				return true, formatConfigSkipReason(entry)
			}
		}
	}

	return false, ""
}

func formatConfigSkipReason(entry ConfigSkipRule) string {
	if strings.TrimSpace(entry.Reason) != "" {
		return entry.Reason
	}
	if strings.TrimSpace(entry.Date) != "" {
		return fmt.Sprintf("Skipped by lint.skip config (%s)", strings.TrimSpace(entry.Date))
	}
	return "Skipped by lint.skip config"
}

func buildSkipPathCandidates(inputFilePath string, modelSourcePath string) []string {
	candidatesMap := map[string]bool{}
	addCandidate := func(value string) {
		normalized := normalizeSkipPath(value)
		if normalized != "" {
			candidatesMap[normalized] = true
		}
	}

	addCandidate(inputFilePath)

	if modelSourcePath != "" {
		if relPath, err := filepath.Rel(modelSourcePath, inputFilePath); err == nil {
			addCandidate(relPath)
		}
	}

	candidates := make([]string, 0, len(candidatesMap)*2)
	for candidate := range candidatesMap {
		candidates = append(candidates, candidate)
		if strings.HasSuffix(candidate, ".yaml") {
			candidates = append(candidates, strings.TrimSuffix(candidate, ".yaml"))
		}
	}

	return candidates
}

func normalizeSkipPath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	normalized = strings.TrimPrefix(normalized, "./")
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "." {
		return ""
	}
	return normalized
}
