package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRuleMetadata_Rego(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("parse valid metadata", func(t *testing.T) {
		regoContent := `# METADATA
# scope: package
# title: Test Rego Rule
# description: A test rule for validation
# authors:
# - Test Author <test@example.com>
# custom:
#  category: Testing
#  rulename: TestRegoRule
#  severity: HIGH
#  rulenumber: "001_0001"
#  remediation: Fix the issue
#  input: .*\.yaml
package test.rule

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains error if {
    not input.Name
    error := "Name is required"
}
`
		regoPath := filepath.Join(tempDir, "valid_metadata.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule, err := parseRuleMetadata_Rego(regoPath)
		if err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		if rule.Title != "Test Rego Rule" {
			t.Errorf("Expected title 'Test Rego Rule', got %q", rule.Title)
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
		if rule.RuleName != "TestRegoRule" {
			t.Errorf("Expected rulename 'TestRegoRule', got %q", rule.RuleName)
		}
		if rule.Pattern != ".*\\.yaml" {
			t.Errorf("Expected pattern '.*\\.yaml', got %q", rule.Pattern)
		}
		if rule.PackageName != "test.rule" {
			t.Errorf("Expected package name 'test.rule', got %q", rule.PackageName)
		}
		if rule.Language != LanguageRego {
			t.Errorf("Expected language 'rego', got %q", rule.Language)
		}
		if rule.Path != regoPath {
			t.Errorf("Expected path %q, got %q", regoPath, rule.Path)
		}
	})

	t.Run("parse metadata with unquoted rulenumber", func(t *testing.T) {
		regoContent := `# METADATA
# scope: package
# title: Unquoted Rulenumber Test
# custom:
#  rulenumber: 002_0002
package test.unquoted

default allow := true
`
		regoPath := filepath.Join(tempDir, "unquoted_rulenumber.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule, err := parseRuleMetadata_Rego(regoPath)
		if err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		if rule.Title != "Unquoted Rulenumber Test" {
			t.Errorf("Expected title 'Unquoted Rulenumber Test', got %q", rule.Title)
		}
		// The rulenumber should be preserved as a string
		if rule.RuleNumber != "002_0002" {
			t.Errorf("Expected rulenumber '002_0002', got %q", rule.RuleNumber)
		}
	})

	t.Run("parse without metadata", func(t *testing.T) {
		regoContent := `package test.no_metadata

default allow := true
`
		regoPath := filepath.Join(tempDir, "no_metadata.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule, err := parseRuleMetadata_Rego(regoPath)
		if err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		if rule.PackageName != "test.no_metadata" {
			t.Errorf("Expected package name 'test.no_metadata', got %q", rule.PackageName)
		}
		if rule.Title != "" {
			t.Errorf("Expected empty title, got %q", rule.Title)
		}
		if rule.Language != LanguageRego {
			t.Errorf("Expected language 'rego', got %q", rule.Language)
		}
	})

	t.Run("parse nonexistent file returns error", func(t *testing.T) {
		_, err := parseRuleMetadata_Rego(filepath.Join(tempDir, "nonexistent.rego"))
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})

	t.Run("parse empty metadata block", func(t *testing.T) {
		regoContent := `# METADATA
package test.empty_metadata

default allow := true
`
		regoPath := filepath.Join(tempDir, "empty_metadata.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rule, err := parseRuleMetadata_Rego(regoPath)
		if err != nil {
			t.Fatalf("Failed to parse metadata: %v", err)
		}

		if rule.PackageName != "test.empty_metadata" {
			t.Errorf("Expected package name 'test.empty_metadata', got %q", rule.PackageName)
		}
	})
}

func TestEvalTestcase_Rego(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("evaluate passing rule", func(t *testing.T) {
		regoContent := `# METADATA
# title: Test Rule
# custom:
#  rulenumber: "001_0001"
package test.pass

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains error if {
    false
    error := "never triggered"
}
`
		regoPath := filepath.Join(tempDir, "pass_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		yamlContent := `Name: "TestEntity"`
		yamlPath := filepath.Join(tempDir, "pass_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Rego(regoPath, "data.test.pass", yamlPath, "001_0001", false)
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
		regoContent := `# METADATA
# title: Test Rule
# custom:
#  rulenumber: "001_0002"
package test.fail

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains error if {
    not input.Name
    error := "Name is required"
}
`
		regoPath := filepath.Join(tempDir, "fail_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		yamlContent := `Value: "NoNameField"`
		yamlPath := filepath.Join(tempDir, "fail_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Rego(regoPath, "data.test.fail", yamlPath, "001_0002", false)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Failure == nil {
			t.Error("Expected failure, got nil")
		} else if testcase.Failure.Message != "Name is required" {
			t.Errorf("Expected message 'Name is required', got %q", testcase.Failure.Message)
		}
		if testcase.Failure.Type != "AssertionError" {
			t.Errorf("Expected type 'AssertionError', got %q", testcase.Failure.Type)
		}
	})

	t.Run("evaluate skipped rule with noqa", func(t *testing.T) {
		regoContent := `# METADATA
# title: Test Rule
# custom:
#  rulenumber: "001_0003"
package test.noqa

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains "Always fails"
`
		regoPath := filepath.Join(tempDir, "noqa_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
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

		testcase, err := evalTestcase_Rego(regoPath, "data.test.noqa", yamlPath, "001_0003", false)
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
		regoContent := `# METADATA
# title: Test Rule
# custom:
#  rulenumber: "001_0004"
package test.ignore_noqa

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains "Always fails"
`
		regoPath := filepath.Join(tempDir, "ignore_noqa_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
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

		testcase, err := evalTestcase_Rego(regoPath, "data.test.ignore_noqa", yamlPath, "001_0004", true)
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
		regoContent := `package test.error

import rego.v1

default allow := true

errors contains error if {
    false
    error := "never triggered"
}
`
		regoPath := filepath.Join(tempDir, "error_test_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		_, err = evalTestcase_Rego(regoPath, "data.test.error", filepath.Join(tempDir, "nonexistent.yaml"), "001_0005", false)
		if err == nil {
			t.Error("Expected error for nonexistent input file")
		}
	})

	t.Run("evaluate rule with multiple errors", func(t *testing.T) {
		regoContent := `# METADATA
# title: Multiple Errors Rule
# custom:
#  rulenumber: "001_0006"
package test.multiple_errors

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains error if {
    not input.Name
    error := "Name is required"
}

errors contains error if {
    not input.Description
    error := "Description is required"
}
`
		regoPath := filepath.Join(tempDir, "multiple_errors_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		yamlContent := `Value: "NoRequiredFields"`
		yamlPath := filepath.Join(tempDir, "multiple_errors_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Rego(regoPath, "data.test.multiple_errors", yamlPath, "001_0006", false)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Failure == nil {
			t.Error("Expected failure, got nil")
		}
		// Multiple errors should be joined with newlines
		if testcase.Failure != nil {
			// Both errors should be present (order may vary)
			msg := testcase.Failure.Message
			hasNameError := containsSubstring(msg, "Name is required")
			hasDescError := containsSubstring(msg, "Description is required")
			if !hasNameError || !hasDescError {
				t.Errorf("Expected both errors in message, got: %q", msg)
			}
		}
	})

	t.Run("evaluate rule with complex input", func(t *testing.T) {
		regoContent := `# METADATA
# title: Complex Input Rule
# custom:
#  rulenumber: "001_0007"
package test.complex

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains error if {
    count(input.Items) < 2
    error := "At least 2 items required"
}
`
		regoPath := filepath.Join(tempDir, "complex_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		yamlContent := `
Items:
  - name: "Item1"
  - name: "Item2"
  - name: "Item3"
`
		yamlPath := filepath.Join(tempDir, "complex_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Rego(regoPath, "data.test.complex", yamlPath, "001_0007", false)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Failure != nil {
			t.Errorf("Expected no failure with 3 items, got: %s", testcase.Failure.Message)
		}
	})

	t.Run("testcase time is recorded", func(t *testing.T) {
		regoContent := `package test.time

import rego.v1

default allow := true

errors contains error if {
    false
    error := "never triggered"
}
`
		regoPath := filepath.Join(tempDir, "time_rule.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		yamlContent := `Name: "Test"`
		yamlPath := filepath.Join(tempDir, "time_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		testcase, err := evalTestcase_Rego(regoPath, "data.test.time", yamlPath, "001_0008", false)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase: %v", err)
		}

		if testcase.Time <= 0 {
			t.Error("Expected positive time value")
		}
	})
}

func TestQuoteRegoMetadataRulenumberIntegration(t *testing.T) {
	// Integration test to ensure rulenumber quoting works correctly during eval
	tempDir := t.TempDir()

	t.Run("unquoted rulenumber with leading zero is handled", func(t *testing.T) {
		// This tests the integration of quoteRegoMetadataRulenumber with evalTestcase_Rego
		regoContent := `# METADATA
# title: Leading Zero Rule
# custom:
#  rulenumber: 002_0001
package test.leading_zero

import rego.v1

default allow := true

errors contains error if {
    false
    error := "never triggered"
}
`
		regoPath := filepath.Join(tempDir, "leading_zero.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		yamlContent := `Name: "Test"`
		yamlPath := filepath.Join(tempDir, "leading_zero_input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		// This should not fail due to YAML 1.1 octal interpretation
		testcase, err := evalTestcase_Rego(regoPath, "data.test.leading_zero", yamlPath, "002_0001", false)
		if err != nil {
			t.Fatalf("Failed to evaluate testcase (possibly rulenumber quoting issue): %v", err)
		}

		if testcase.Failure != nil {
			t.Errorf("Expected no failure, got: %s", testcase.Failure.Message)
		}
	})
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
