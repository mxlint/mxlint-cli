package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func setDefaultConfigForTest(t *testing.T, configYAML string) {
	t.Helper()
	SetDefaultConfigYAML([]byte(configYAML))
	t.Cleanup(func() {
		SetDefaultConfigYAML(nil)
	})
}

func TestLoadMergedConfig_ProjectOverridesSystem(t *testing.T) {
	systemDir := t.TempDir()
	projectDir := t.TempDir()
	systemConfigPath := filepath.Join(systemDir, "mxlint.yaml")
	setDefaultConfigForTest(t, "")

	systemConfig := `rules:
  path: .mendix-cache/system-rules
  rulesets:
    - file://system-rules
modelsource: modelsource-system
projectDirectory: ./system
export:
  mode: advanced
  filter: system/*
lint:
  skip:
    example/doc:
      - rule: 001_002
        reason: system reason
`
	projectConfig := `rules:
  path: .mendix-cache/project-rules
  rulesets:
    - file://project-rules
modelsource: modelsource
projectDirectory: .
export:
  mode: basic
  filter: "*"
lint:
  skip:
    example/doc:
      - rule: 001_002
        reason: project reason
    project/only:
      - rule: 003_004
        reason: project only
`

	if err := os.WriteFile(systemConfigPath, []byte(systemConfig), 0644); err != nil {
		t.Fatalf("failed to write system config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "mxlint.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	original := os.Getenv("MXLINT_SYSTEM_CONFIG")
	t.Setenv("MXLINT_SYSTEM_CONFIG", systemConfigPath)
	defer func() {
		_ = os.Setenv("MXLINT_SYSTEM_CONFIG", original)
	}()

	cfg, err := LoadMergedConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadMergedConfig returned error: %v", err)
	}

	if cfg.Rules.Path != ".mendix-cache/project-rules" {
		t.Fatalf("expected project rules.path override, got %s", cfg.Rules.Path)
	}
	if len(cfg.Rules.Rulesets) != 1 || cfg.Rules.Rulesets[0] != "file://project-rules" {
		t.Fatalf("expected project rules.rulesets override, got %#v", cfg.Rules.Rulesets)
	}
	if cfg.Modelsource != "modelsource" || cfg.ProjectDirectory != "." || cfg.Export.Mode != "basic" || cfg.Export.Filter != "*" {
		t.Fatalf("expected project config override, got modelsource=%s projectDirectory=%s export=%#v", cfg.Modelsource, cfg.ProjectDirectory, cfg.Export)
	}

	entry := cfg.Lint.Skip["example/doc"][0]
	if entry.Reason != "project reason" {
		t.Fatalf("expected project skip reason override, got %s", entry.Reason)
	}
	if _, ok := cfg.Lint.Skip["project/only"]; !ok {
		t.Fatal("expected project-only skip entry to exist")
	}
}

func TestShouldSkipRule_ConfigSkipApplied(t *testing.T) {
	setDefaultConfigForTest(t, "")
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	SetConfig(&Config{
		Lint: ConfigLintSpec{
			Skip: map[string][]ConfigSkipRule{
				"example/doc": []ConfigSkipRule{
					{
						Rule:   "001_002",
						Reason: "skip from config",
					},
				},
			},
		},
	})

	skip, reason := shouldSkipRule("", "001_002", true, "/tmp/modelsource/example/doc.yaml", "/tmp/modelsource")
	if !skip {
		t.Fatal("expected skip=true for configured skip entry")
	}
	if reason != "skip from config" {
		t.Fatalf("expected config reason, got %s", reason)
	}
}

func TestLoadMergedConfig_DefaultConfigBase(t *testing.T) {
	projectDir := t.TempDir()
	defaultConfig := `rules:
  path: .mendix-cache/default-rules
modelsource: default-modelsource
projectDirectory: ./default-input
export:
  mode: advanced
  filter: default/*
`
	setDefaultConfigForTest(t, defaultConfig)

	cfg, err := LoadMergedConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadMergedConfig returned error: %v", err)
	}

	if cfg.Rules.Path != ".mendix-cache/default-rules" {
		t.Fatalf("expected default rules path, got %s", cfg.Rules.Path)
	}
	if cfg.Modelsource != "default-modelsource" || cfg.ProjectDirectory != "./default-input" || cfg.Export.Mode != "advanced" || cfg.Export.Filter != "default/*" {
		t.Fatalf("expected default values, got modelsource=%s projectDirectory=%s export=%#v", cfg.Modelsource, cfg.ProjectDirectory, cfg.Export)
	}
}

func TestLoadMergedConfig_UnquotedRuleNumber(t *testing.T) {
	projectDir := t.TempDir()
	setDefaultConfigForTest(t, "")
	projectConfig := `lint:
  skip:
    example/doc:
      - rule: 001_002
        reason: unquoted rule
`
	if err := os.WriteFile(filepath.Join(projectDir, "mxlint.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := LoadMergedConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadMergedConfig returned error: %v", err)
	}
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	skip, reason := shouldSkipRule("", "001_002", true, "/tmp/modelsource/example/doc.yaml", "/tmp/modelsource")
	if !skip {
		t.Fatal("expected lint.skip to match unquoted rule number")
	}
	if reason != "unquoted rule" {
		t.Fatalf("expected configured reason, got %s", reason)
	}
}

func TestLoadMergedConfigWithReport_Sources(t *testing.T) {
	projectDir := t.TempDir()
	systemDir := t.TempDir()
	systemConfigPath := filepath.Join(systemDir, "mxlint.yaml")

	setDefaultConfigForTest(t, "rules:\n  path: embedded-rules\n")
	if err := os.WriteFile(systemConfigPath, []byte("rules:\n  path: system-rules\n"), 0644); err != nil {
		t.Fatalf("failed to write system config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "mxlint.yaml"), []byte("rules:\n  path: project-rules\n"), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}
	t.Setenv("MXLINT_SYSTEM_CONFIG", systemConfigPath)

	cfg, report, err := LoadMergedConfigWithReport(projectDir)
	if err != nil {
		t.Fatalf("LoadMergedConfigWithReport returned error: %v", err)
	}

	if !report.Default.Found || !report.Default.Used {
		t.Fatalf("expected embedded default source found+used, got %#v", report.Default)
	}
	if !report.System.Found || !report.System.Used {
		t.Fatalf("expected system source found+used, got %#v", report.System)
	}
	if !report.Project.Found || !report.Project.Used {
		t.Fatalf("expected project source found+used, got %#v", report.Project)
	}
	if report.Explicit.Found || report.Explicit.Used {
		t.Fatalf("expected explicit source not found+not used by default, got %#v", report.Explicit)
	}
	if cfg.Rules.Path != "project-rules" {
		t.Fatalf("expected merged project value to win, got %s", cfg.Rules.Path)
	}
}

func TestLoadMergedConfigFromPath_ExplicitOverridesProject(t *testing.T) {
	projectDir := t.TempDir()
	setDefaultConfigForTest(t, "")

	projectConfig := "rules:\n  path: project-rules\n"
	explicitConfig := "rules:\n  path: explicit-rules\n"
	explicitPath := filepath.Join(projectDir, "custom.yaml")
	if err := os.WriteFile(filepath.Join(projectDir, "mxlint.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}
	if err := os.WriteFile(explicitPath, []byte(explicitConfig), 0644); err != nil {
		t.Fatalf("failed to write explicit config: %v", err)
	}

	cfg, report, err := LoadMergedConfigWithReportFromPath(projectDir, "custom.yaml")
	if err != nil {
		t.Fatalf("LoadMergedConfigWithReportFromPath returned error: %v", err)
	}
	if cfg.Rules.Path != "explicit-rules" {
		t.Fatalf("expected explicit config to win, got %s", cfg.Rules.Path)
	}
	if !report.Explicit.Found || !report.Explicit.Used {
		t.Fatalf("expected explicit source found+used, got %#v", report.Explicit)
	}
}

func TestLoadMergedConfigFromPath_MissingExplicitReturnsError(t *testing.T) {
	projectDir := t.TempDir()
	setDefaultConfigForTest(t, "")

	_, err := LoadMergedConfigFromPath(projectDir, "missing.yaml")
	if err == nil {
		t.Fatal("expected error when explicit config file is missing")
	}
}

func TestShouldSkipRule_ConfigSkipPathVariants(t *testing.T) {
	setDefaultConfigForTest(t, "")
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	SetConfig(&Config{
		Lint: ConfigLintSpec{
			Skip: map[string][]ConfigSkipRule{
				"example/doc": []ConfigSkipRule{
					{
						Rule:   "001_002",
						Reason: "normalized path skip",
					},
				},
			},
		},
	})

	inputFile := "/tmp/modelsource/./example/doc.yaml"
	skip, reason := shouldSkipRule("", "001_002", false, inputFile, "/tmp/modelsource")
	if !skip {
		t.Fatal("expected skip=true for normalized path candidate")
	}
	if reason != "normalized path skip" {
		t.Fatalf("expected configured reason, got %s", reason)
	}
}

func TestShouldSkipRule_ConfigSkipWildcardRuleWithDateReason(t *testing.T) {
	setDefaultConfigForTest(t, "")
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	SetConfig(&Config{
		Lint: ConfigLintSpec{
			Skip: map[string][]ConfigSkipRule{
				"example/doc": []ConfigSkipRule{
					{
						Rule: "*",
						Date: "2026-03-03",
					},
				},
			},
		},
	})

	skip, reason := shouldSkipRule("", "009_9999", false, "/tmp/modelsource/example/doc.yaml", "/tmp/modelsource")
	if !skip {
		t.Fatal("expected skip=true for wildcard rule entry")
	}
	if reason != "Skipped by lint.skip config (2026-03-03)" {
		t.Fatalf("expected date-based skip reason, got %s", reason)
	}
}

func TestLoadMergedConfig_NormalizesSkipMapKeys(t *testing.T) {
	projectDir := t.TempDir()
	setDefaultConfigForTest(t, "")
	projectConfig := `lint:
  skip:
    ./example/doc:
      - rule: "001_002"
        reason: normalized
`
	if err := os.WriteFile(filepath.Join(projectDir, "mxlint.yaml"), []byte(projectConfig), 0644); err != nil {
		t.Fatalf("failed to write project config: %v", err)
	}

	cfg, err := LoadMergedConfig(projectDir)
	if err != nil {
		t.Fatalf("LoadMergedConfig returned error: %v", err)
	}

	if _, ok := cfg.Lint.Skip["example/doc"]; !ok {
		t.Fatalf("expected normalized skip key example/doc, got %#v", cfg.Lint.Skip)
	}
	if _, ok := cfg.Lint.Skip["./example/doc"]; ok {
		t.Fatalf("unexpected unnormalized skip key present: %#v", cfg.Lint.Skip)
	}
}
