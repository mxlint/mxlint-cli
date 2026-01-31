package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashRuleContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected string
	}{
		{
			name:     "Empty content",
			content:  []byte(""),
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Simple content",
			content:  []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "Same content produces same hash",
			content:  []byte("test content"),
			expected: "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashRuleContent(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTranspileTypescriptRule(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("transpile valid TypeScript", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0001" }
};

function rule(input: { Name?: string }): { allow: boolean; errors: string[] } {
    const errors: string[] = [];
    return { allow: true, errors };
}
`
		tsPath := filepath.Join(tempDir, "test_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		result, err := transpileTypescriptRule(tsPath)
		if err != nil {
			t.Fatalf("Failed to transpile TypeScript: %v", err)
		}

		// Check that TypeScript type annotations are removed
		if result == "" {
			t.Error("Expected non-empty transpiled code")
		}

		// TypeScript types should be removed
		if containsString(result, ": string[]") {
			t.Error("Expected TypeScript type annotations to be removed")
		}
	})

	t.Run("transpile caches result", func(t *testing.T) {
		tsContent := `const x: number = 1;`
		tsPath := filepath.Join(tempDir, "cached_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		// First call
		result1, err := transpileTypescriptRule(tsPath)
		if err != nil {
			t.Fatalf("First transpile failed: %v", err)
		}

		// Second call should use cache
		result2, err := transpileTypescriptRule(tsPath)
		if err != nil {
			t.Fatalf("Second transpile failed: %v", err)
		}

		if result1 != result2 {
			t.Error("Expected cached result to match original result")
		}
	})

	t.Run("cache invalidated on content change", func(t *testing.T) {
		tsPath := filepath.Join(tempDir, "changing_rule.ts")

		// Write initial content
		err := os.WriteFile(tsPath, []byte(`const x: number = 1;`), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		result1, err := transpileTypescriptRule(tsPath)
		if err != nil {
			t.Fatalf("First transpile failed: %v", err)
		}

		// Modify the file
		err = os.WriteFile(tsPath, []byte(`const y: number = 2;`), 0644)
		if err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		result2, err := transpileTypescriptRule(tsPath)
		if err != nil {
			t.Fatalf("Second transpile failed: %v", err)
		}

		if result1 == result2 {
			t.Error("Expected different result after content change")
		}
	})

	t.Run("transpile nonexistent file returns error", func(t *testing.T) {
		_, err := transpileTypescriptRule(filepath.Join(tempDir, "nonexistent.ts"))
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("transpile invalid TypeScript returns error", func(t *testing.T) {
		tsContent := `
function rule(input { // missing colon - syntax error
    return { allow: true };
}
`
		tsPath := filepath.Join(tempDir, "invalid_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err = transpileTypescriptRule(tsPath)
		if err == nil {
			t.Error("Expected error for invalid TypeScript")
		}
	})
}

func TestParseRuleMetadata_Typescript(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("parse valid metadata", func(t *testing.T) {
		tsContent := `
const metadata = {
    scope: "package",
    title: "Test TypeScript Rule",
    description: "A test rule for validation",
    authors: ["Test Author <test@example.com>"],
    custom: {
        category: "Testing",
        rulename: "TestTypescriptRule",
        severity: "HIGH",
        rulenumber: "001_0001",
        remediation: "Fix the issue",
        input: ".*\\.yaml"
    }
};

function rule(input) {
    return { allow: true, errors: [] };
}
`
		tsPath := filepath.Join(tempDir, "valid_metadata.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule, err := parseRuleMetadata_Typescript(tsPath)
		if err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		if rule.Title != "Test TypeScript Rule" {
			t.Errorf("Expected title 'Test TypeScript Rule', got %q", rule.Title)
		}
		if rule.Description != "A test rule for validation" {
			t.Errorf("Expected description 'A test rule for validation', got %q", rule.Description)
		}
		if rule.Category != "Testing" {
			t.Errorf("Expected category 'Testing', got %q", rule.Category)
		}
		if rule.Severity != "HIGH" {
			t.Errorf("Expected severity 'HIGH', got %q", rule.Severity)
		}
		if rule.RuleNumber != "001_0001" {
			t.Errorf("Expected rulenumber '001_0001', got %q", rule.RuleNumber)
		}
		if rule.Remediation != "Fix the issue" {
			t.Errorf("Expected remediation 'Fix the issue', got %q", rule.Remediation)
		}
		if rule.RuleName != "TestTypescriptRule" {
			t.Errorf("Expected rulename 'TestTypescriptRule', got %q", rule.RuleName)
		}
		if rule.Pattern != ".*\\.yaml" {
			t.Errorf("Expected pattern '.*\\.yaml', got %q", rule.Pattern)
		}
		if rule.Language != LanguageTypescript {
			t.Errorf("Expected language 'typescript', got %q", rule.Language)
		}
		if rule.Path != tsPath {
			t.Errorf("Expected path %q, got %q", tsPath, rule.Path)
		}
	})
}

func TestEvalTestcase_Typescript(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("evaluate passing rule", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0001", input: ".*\\.yaml" }
};

function rule(input) {
    const errors = [];
    if (!input.Name) {
        errors.push("Name is required");
    }
    return { allow: errors.length === 0, errors };
}
`
		tsPath := filepath.Join(tempDir, "pass_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		yamlContent := `Name: "TestEntity"`
		yamlPath := filepath.Join(tempDir, "pass_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Typescript(tsPath, yamlPath, "001_0001", false, tempDir)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Failure != nil {
			t.Errorf("Expected no failure, got: %s", testcase.Failure.Message)
		}
		if testcase.Name != yamlPath {
			t.Errorf("Expected name %q, got %q", yamlPath, testcase.Name)
		}
	})

	t.Run("evaluate failing rule", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0002", input: ".*\\.yaml" }
};

function rule(input) {
    const errors = [];
    if (!input.Name) {
        errors.push("Name is required");
    }
    return { allow: errors.length === 0, errors };
}
`
		tsPath := filepath.Join(tempDir, "fail_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		yamlContent := `Value: "NoNameField"`
		yamlPath := filepath.Join(tempDir, "fail_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Typescript(tsPath, yamlPath, "001_0002", false, tempDir)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Failure == nil {
			t.Error("Expected failure, got nil")
		} else if testcase.Failure.Message != "Name is required" {
			t.Errorf("Expected message 'Name is required', got %q", testcase.Failure.Message)
		}
	})

	t.Run("evaluate skipped rule with noqa", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0003", input: ".*\\.yaml" }
};

function rule(input) {
    return { allow: false, errors: ["Always fails"] };
}
`
		tsPath := filepath.Join(tempDir, "noqa_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		yamlContent := `
Documentation: "#noqa:001_0003"
Name: "Test"
`
		yamlPath := filepath.Join(tempDir, "noqa_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Typescript(tsPath, yamlPath, "001_0003", false, tempDir)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Skipped == nil {
			t.Error("Expected testcase to be skipped")
		}
		if testcase.Failure != nil {
			t.Error("Expected no failure when skipped")
		}
	})

	t.Run("noqa ignored when ignoreNoqa is true", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0004", input: ".*\\.yaml" }
};

function rule(input) {
    return { allow: false, errors: ["Always fails"] };
}
`
		tsPath := filepath.Join(tempDir, "ignore_noqa_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		yamlContent := `
Documentation: "#noqa:001_0004"
Name: "Test"
`
		yamlPath := filepath.Join(tempDir, "ignore_noqa_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Typescript(tsPath, yamlPath, "001_0004", true, tempDir)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Skipped != nil {
			t.Error("Expected testcase not to be skipped when ignoreNoqa is true")
		}
		if testcase.Failure == nil {
			t.Error("Expected failure when ignoreNoqa is true")
		}
	})

	t.Run("error reading nonexistent input file", func(t *testing.T) {
		tsContent := `
const metadata = { title: "Test", custom: { rulenumber: "001_0005" } };
function rule(input) { return { allow: true, errors: [] }; }
`
		tsPath := filepath.Join(tempDir, "error_test_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		_, err = evalTestcase_Typescript(tsPath, filepath.Join(tempDir, "nonexistent.yaml"), "001_0005", false, tempDir)
		if err == nil {
			t.Error("Expected error for nonexistent input file")
		}
	})
}

func TestRunTypescriptTestCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("run passing test cases", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0001", input: ".*\\.yaml" }
};

function rule(input) {
    const errors = [];
    if (!input.Name) {
        errors.push("Name is required");
    }
    return { allow: errors.length === 0, errors };
}
`
		tsPath := filepath.Join(tempDir, "testcase_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		testYamlContent := `
TestCases:
  - name: "Test with name present"
    input:
      Name: "TestName"
    allow: true
  - name: "Test with name missing"
    input:
      Value: "SomeValue"
    allow: false
`
		testYamlPath := filepath.Join(tempDir, "testcase_rule_test.yaml")
		err = os.WriteFile(testYamlPath, []byte(testYamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test yaml file: %v", err)
		}

		rule := Rule{
			Path:     tsPath,
			Language: LanguageTypescript,
		}

		err = runTypescriptTestCases(rule)
		if err != nil {
			t.Errorf("Expected test cases to pass, got error: %v", err)
		}
	})

	t.Run("run failing test cases", func(t *testing.T) {
		tsContent := `
const metadata = {
    title: "Test Rule",
    custom: { rulenumber: "001_0002", input: ".*\\.yaml" }
};

function rule(input) {
    return { allow: true, errors: [] };
}
`
		tsPath := filepath.Join(tempDir, "failing_testcase_rule.ts")
		err := os.WriteFile(tsPath, []byte(tsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		testYamlContent := `
TestCases:
  - name: "This should fail because rule always allows but test expects deny"
    input:
      Name: "TestName"
    allow: false
`
		testYamlPath := filepath.Join(tempDir, "failing_testcase_rule_test.yaml")
		err = os.WriteFile(testYamlPath, []byte(testYamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test yaml file: %v", err)
		}

		rule := Rule{
			Path:     tsPath,
			Language: LanguageTypescript,
		}

		err = runTypescriptTestCases(rule)
		if err == nil {
			t.Error("Expected test cases to fail")
		}
	})

	t.Run("test file not found", func(t *testing.T) {
		tsPath := filepath.Join(tempDir, "no_test_file.ts")
		err := os.WriteFile(tsPath, []byte(`const metadata = {}; function rule(input) { return { allow: true, errors: [] }; }`), 0644)
		if err != nil {
			t.Fatalf("Failed to write rule file: %v", err)
		}

		rule := Rule{
			Path:     tsPath,
			Language: LanguageTypescript,
		}

		err = runTypescriptTestCases(rule)
		if err == nil {
			t.Error("Expected error when test file is missing")
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
