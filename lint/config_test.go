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
export:
  output: modelsource-system
  input: ./system
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
export:
  output: modelsource
  input: .
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
	if cfg.Export.Output != "modelsource" || cfg.Export.Input != "." || cfg.Export.Mode != "basic" || cfg.Export.Filter != "*" {
		t.Fatalf("expected project export override, got %#v", cfg.Export)
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
export:
  output: default-modelsource
  input: ./default-input
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
	if cfg.Export.Output != "default-modelsource" || cfg.Export.Input != "./default-input" || cfg.Export.Mode != "advanced" || cfg.Export.Filter != "default/*" {
		t.Fatalf("expected default export values, got %#v", cfg.Export)
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
