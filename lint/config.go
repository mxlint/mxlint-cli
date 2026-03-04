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
	Rules            ConfigRulesSpec  `yaml:"rules"`
	Lint             ConfigLintSpec   `yaml:"lint"`
	Export           ConfigExportSpec `yaml:"export"`
	Serve            ConfigServeSpec  `yaml:"serve"`
	Modelsource      string           `yaml:"modelsource"`
	ProjectDirectory string           `yaml:"projectDirectory"`
}

type ConfigRulesSpec struct {
	Path     string   `yaml:"path"`
	Rulesets []string `yaml:"rulesets"`
}

type ConfigExportSpec struct {
	Mode     string `yaml:"mode"`
	Filter   string `yaml:"filter"`
	Raw      *bool  `yaml:"raw"`
	Appstore *bool  `yaml:"appstore"`
}

type ConfigLintSpec struct {
	XunitReport string                      `yaml:"xunitReport"`
	JSONFile    string                      `yaml:"jsonFile"`
	IgnoreNoqa  *bool                       `yaml:"ignoreNoqa"`
	NoCache     *bool                       `yaml:"noCache"`
	Skip        map[string][]ConfigSkipRule `yaml:"skip"`
}

type ConfigServeSpec struct {
	Port     *int `yaml:"port"`
	Debounce *int `yaml:"debounce"`
}

type ConfigSkipRule struct {
	Rule   string `yaml:"rule"`
	Reason string `yaml:"reason"`
	Date   string `yaml:"date"`
}

type ConfigSourceStatus struct {
	Name  string
	Path  string
	Found bool
	Used  bool
}

type ConfigLoadReport struct {
	Default  ConfigSourceStatus
	System   ConfigSourceStatus
	Project  ConfigSourceStatus
	Explicit ConfigSourceStatus
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
	cfg, _, err := LoadMergedConfigWithReport(projectDir)
	return cfg, err
}

func LoadMergedConfigFromPath(projectDir string, explicitConfigPath string) (*Config, error) {
	cfg, _, err := LoadMergedConfigWithReportFromPath(projectDir, explicitConfigPath)
	return cfg, err
}

func LoadMergedConfigWithReport(projectDir string) (*Config, ConfigLoadReport, error) {
	return LoadMergedConfigWithReportFromPath(projectDir, "")
}

func LoadMergedConfigWithReportFromPath(projectDir string, explicitConfigPath string) (*Config, ConfigLoadReport, error) {
	cfg := &Config{}
	report := ConfigLoadReport{
		Default:  ConfigSourceStatus{Name: "embedded-default", Path: "embedded:default.yaml"},
		System:   ConfigSourceStatus{Name: "system", Path: ""},
		Project:  ConfigSourceStatus{Name: "project", Path: filepath.Join(projectDir, configFileName)},
		Explicit: ConfigSourceStatus{Name: "explicit", Path: strings.TrimSpace(explicitConfigPath)},
	}

	systemConfigPath, err := resolveSystemConfigPath()
	if err != nil {
		return nil, report, err
	}
	report.System.Path = systemConfigPath
	projectConfigPath := filepath.Join(projectDir, configFileName)
	resolvedExplicitPath := strings.TrimSpace(explicitConfigPath)
	if resolvedExplicitPath != "" && !filepath.IsAbs(resolvedExplicitPath) {
		resolvedExplicitPath = filepath.Join(projectDir, resolvedExplicitPath)
	}
	report.Explicit.Path = resolvedExplicitPath

	defaultCfg, defaultFound, err := loadConfigYAML(getDefaultConfigYAML(), "embedded default config")
	if err != nil {
		return nil, report, err
	}
	report.Default.Found = defaultFound
	report.Default.Used = defaultCfg != nil
	mergeConfig(cfg, defaultCfg)

	systemCfg, systemFound, err := loadConfigFile(systemConfigPath)
	if err != nil {
		return nil, report, err
	}
	report.System.Found = systemFound
	report.System.Used = systemCfg != nil
	mergeConfig(cfg, systemCfg)

	projectCfg, projectFound, err := loadConfigFile(projectConfigPath)
	if err != nil {
		return nil, report, err
	}
	report.Project.Found = projectFound
	report.Project.Used = projectCfg != nil
	mergeConfig(cfg, projectCfg)

	explicitCfg, explicitFound, err := loadConfigFileRequired(resolvedExplicitPath)
	if err != nil {
		return nil, report, err
	}
	report.Explicit.Found = explicitFound
	report.Explicit.Used = explicitCfg != nil
	mergeConfig(cfg, explicitCfg)

	return cfg, report, nil
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

func loadConfigFile(configPath string) (*Config, bool, error) {
	if configPath == "" {
		return nil, false, nil
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}
	cfg, _, err := loadConfigYAML(content, configPath)
	if err != nil {
		return nil, true, err
	}
	return cfg, true, nil
}

func loadConfigFileRequired(configPath string) (*Config, bool, error) {
	if configPath == "" {
		return nil, false, nil
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, fmt.Errorf("failed to read config %s: file does not exist", configPath)
		}
		return nil, false, fmt.Errorf("failed to read config %s: %w", configPath, err)
	}
	cfg, _, err := loadConfigYAML(content, configPath)
	if err != nil {
		return nil, true, err
	}
	return cfg, true, nil
}

func loadConfigYAML(content []byte, source string) (*Config, bool, error) {
	if len(content) == 0 {
		return nil, false, nil
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, true, fmt.Errorf("failed to parse config %s: %w", source, err)
	}
	return cfg, true, nil
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

	if strings.TrimSpace(overlay.Export.Mode) != "" {
		base.Export.Mode = strings.TrimSpace(overlay.Export.Mode)
	}
	if overlay.Export.Filter != "" {
		base.Export.Filter = strings.TrimSpace(overlay.Export.Filter)
	}
	if overlay.Export.Raw != nil {
		base.Export.Raw = overlay.Export.Raw
	}
	if overlay.Export.Appstore != nil {
		base.Export.Appstore = overlay.Export.Appstore
	}

	if strings.TrimSpace(overlay.Modelsource) != "" {
		base.Modelsource = strings.TrimSpace(overlay.Modelsource)
	}
	if strings.TrimSpace(overlay.ProjectDirectory) != "" {
		base.ProjectDirectory = strings.TrimSpace(overlay.ProjectDirectory)
	}
	if overlay.Lint.XunitReport != "" {
		base.Lint.XunitReport = strings.TrimSpace(overlay.Lint.XunitReport)
	}
	if overlay.Lint.JSONFile != "" {
		base.Lint.JSONFile = strings.TrimSpace(overlay.Lint.JSONFile)
	}
	if overlay.Lint.IgnoreNoqa != nil {
		base.Lint.IgnoreNoqa = overlay.Lint.IgnoreNoqa
	}
	if overlay.Lint.NoCache != nil {
		base.Lint.NoCache = overlay.Lint.NoCache
	}

	if overlay.Serve.Port != nil {
		base.Serve.Port = overlay.Serve.Port
	}
	if overlay.Serve.Debounce != nil {
		base.Serve.Debounce = overlay.Serve.Debounce
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
