package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvalTestsuite_Rego(t *testing.T) {
	t.Run("single Rego rule passes", func(t *testing.T) {
		rule, err := parseRuleMetadata_Rego("./../resources/rules/001_0003_security_checks.rego")
		if err != nil {
			t.Fatalf("Failed to parse rule metadata: %v", err)
		}

		result, err := evalTestsuite(*rule, "./../resources/modelsource-v1", false, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Failures != 0 {
			t.Errorf("Expected no failures, got %d", result.Failures)
		}
		if result.Tests == 0 {
			t.Error("Expected at least one test case")
		}
	})

	t.Run("Rego rule with failures", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a rule that always fails
		regoContent := `# METADATA
# title: Always Fail Rule
# custom:
#  rulenumber: "099_0001"
#  input: .*\.yaml
package test.always_fail

import rego.v1

default allow := false
allow if count(errors) == 0

errors contains "Always fails"
`
		regoPath := filepath.Join(tempDir, "always_fail.rego")
		err := os.WriteFile(regoPath, []byte(regoContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write rego file: %v", err)
		}

		// Create a test input file
		yamlContent := `Name: "Test"`
		yamlPath := filepath.Join(tempDir, "input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		rule := Rule{
			Path:        regoPath,
			Pattern:     ".*\\.yaml",
			PackageName: "test.always_fail",
			RuleNumber:  "099_0001",
			Language:    LanguageRego,
		}

		result, err := evalTestsuite(rule, tempDir, false, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Failures != 1 {
			t.Errorf("Expected 1 failure, got %d", result.Failures)
		}
	})
}

func TestEvalTestsuite_Javascript(t *testing.T) {
	t.Run("single JS rule passes", func(t *testing.T) {
		rule, err := parseRuleMetadata_Javascript("./../resources/rules/001_0002_demo_users_disabled.js")
		if err != nil {
			t.Fatalf("Failed to parse rule metadata: %v", err)
		}

		result, err := evalTestsuite(*rule, "./../resources/modelsource-v1", false, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Failures != 0 {
			t.Errorf("Expected no failures, got %d", result.Failures)
		}
		if result.Tests == 0 {
			t.Error("Expected at least one test case")
		}
	})

	t.Run("JS rule with failures", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a rule that always fails
		jsContent := `
const metadata = {
    title: "Always Fail Rule",
    custom: { rulenumber: "099_0002", input: ".*\\.yaml" }
};

function rule(input) {
    return { allow: false, errors: ["Always fails"] };
}
`
		jsPath := filepath.Join(tempDir, "always_fail.js")
		err := os.WriteFile(jsPath, []byte(jsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write js file: %v", err)
		}

		// Create a test input file
		yamlContent := `Name: "Test"`
		yamlPath := filepath.Join(tempDir, "input.yaml")
		err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		rule := Rule{
			Path:        jsPath,
			Pattern:     ".*\\.yaml",
			PackageName: jsPath,
			RuleNumber:  "099_0002",
			Language:    LanguageJavascript,
		}

		result, err := evalTestsuite(rule, tempDir, false, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Failures != 1 {
			t.Errorf("Expected 1 failure, got %d", result.Failures)
		}
	})
}

func TestEvalTestsuite_Typescript(t *testing.T) {
	t.Run("single TS rule passes", func(t *testing.T) {
		rule, err := parseRuleMetadata_Typescript("./../resources/rules/001_0005_typescript_example.ts")
		if err != nil {
			t.Fatalf("Failed to parse rule metadata: %v", err)
		}

		result, err := evalTestsuite(*rule, "./../resources/modelsource-v1", false, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Failures != 0 {
			t.Errorf("Expected no failures, got %d", result.Failures)
		}
	})
}

func TestEvalTestsuite_WithNoqa(t *testing.T) {
	tempDir := t.TempDir()

	// Create a rule that always fails
	jsContent := `
const metadata = {
    title: "Noqa Test Rule",
    custom: { rulenumber: "099_0003", input: ".*\\.yaml" }
};

function rule(input) {
    return { allow: false, errors: ["Should be skipped"] };
}
`
	jsPath := filepath.Join(tempDir, "noqa_test.js")
	err := os.WriteFile(jsPath, []byte(jsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write js file: %v", err)
	}

	// Create input file with noqa directive
	yamlContent := `
Documentation: "#noqa:099_0003"
Name: "Test"
`
	yamlPath := filepath.Join(tempDir, "noqa_input.yaml")
	err = os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml file: %v", err)
	}

	rule := Rule{
		Path:        jsPath,
		Pattern:     ".*\\.yaml",
		PackageName: jsPath,
		RuleNumber:  "099_0003",
		Language:    LanguageJavascript,
	}

	t.Run("noqa skips the rule", func(t *testing.T) {
		result, err := evalTestsuite(rule, tempDir, false, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Skipped != 1 {
			t.Errorf("Expected 1 skipped, got %d", result.Skipped)
		}
		if result.Failures != 0 {
			t.Errorf("Expected 0 failures when skipped, got %d", result.Failures)
		}
	})

	t.Run("ignoreNoqa runs the rule anyway", func(t *testing.T) {
		result, err := evalTestsuite(rule, tempDir, true, false)
		if err != nil {
			t.Fatalf("Failed to evaluate testsuite: %v", err)
		}

		if result.Skipped != 0 {
			t.Errorf("Expected 0 skipped when ignoreNoqa=true, got %d", result.Skipped)
		}
		if result.Failures != 1 {
			t.Errorf("Expected 1 failure when ignoreNoqa=true, got %d", result.Failures)
		}
	})
}

func TestEvalTestsuite_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a rule that checks for Name field
	jsContent := `
const metadata = {
    title: "Name Required Rule",
    custom: { rulenumber: "099_0004", input: ".*\\.yaml" }
};

function rule(input) {
    const errors = [];
    if (!input.Name) {
        errors.push("Name is required");
    }
    return { allow: errors.length === 0, errors };
}
`
	jsPath := filepath.Join(tempDir, "name_required.js")
	err := os.WriteFile(jsPath, []byte(jsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write js file: %v", err)
	}

	// Create passing input file
	err = os.WriteFile(filepath.Join(tempDir, "pass.yaml"), []byte(`Name: "Test"`), 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml file: %v", err)
	}

	// Create failing input file
	err = os.WriteFile(filepath.Join(tempDir, "fail.yaml"), []byte(`Value: "NoName"`), 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml file: %v", err)
	}

	rule := Rule{
		Path:        jsPath,
		Pattern:     ".*\\.yaml",
		PackageName: jsPath,
		RuleNumber:  "099_0004",
		Language:    LanguageJavascript,
	}

	result, err := evalTestsuite(rule, tempDir, false, false)
	if err != nil {
		t.Fatalf("Failed to evaluate testsuite: %v", err)
	}

	if result.Tests != 2 {
		t.Errorf("Expected 2 tests, got %d", result.Tests)
	}
	if result.Failures != 1 {
		t.Errorf("Expected 1 failure, got %d", result.Failures)
	}
}

func TestCountTotalTestcases(t *testing.T) {
	tests := []struct {
		name       string
		testsuites []Testsuite
		expected   int
	}{
		{
			name:       "Empty testsuites",
			testsuites: []Testsuite{},
			expected:   0,
		},
		{
			name: "Single testsuite with one testcase",
			testsuites: []Testsuite{
				{Testcases: []Testcase{{Name: "test1"}}},
			},
			expected: 1,
		},
		{
			name: "Multiple testsuites",
			testsuites: []Testsuite{
				{Testcases: []Testcase{{Name: "test1"}, {Name: "test2"}}},
				{Testcases: []Testcase{{Name: "test3"}}},
				{Testcases: []Testcase{{Name: "test4"}, {Name: "test5"}, {Name: "test6"}}},
			},
			expected: 6,
		},
		{
			name: "Testsuite with no testcases",
			testsuites: []Testsuite{
				{Testcases: []Testcase{}},
				{Testcases: []Testcase{{Name: "test1"}}},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countTotalTestcases(tt.testsuites)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestReadRulesMetadata(t *testing.T) {
	t.Run("read rules from resources directory", func(t *testing.T) {
		rules, err := ReadRulesMetadata("./../resources/rules")
		if err != nil {
			t.Fatalf("Failed to read rules metadata: %v", err)
		}

		if len(rules) == 0 {
			t.Error("Expected at least one rule")
		}

		// Verify we have different languages
		hasRego := false
		hasJS := false
		hasTS := false
		for _, rule := range rules {
			switch rule.Language {
			case LanguageRego:
				hasRego = true
			case LanguageJavascript:
				hasJS = true
			case LanguageTypescript:
				hasTS = true
			}
		}

		if !hasRego {
			t.Error("Expected at least one Rego rule")
		}
		if !hasJS {
			t.Error("Expected at least one JavaScript rule")
		}
		if !hasTS {
			t.Error("Expected at least one TypeScript rule")
		}
	})

	t.Run("read from empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		rules, err := ReadRulesMetadata(tempDir)
		if err != nil {
			t.Fatalf("Failed to read rules metadata: %v", err)
		}

		if len(rules) != 0 {
			t.Errorf("Expected 0 rules from empty directory, got %d", len(rules))
		}
	})

	t.Run("ignores test files", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a regular rule with proper metadata structure
		jsContent := `
const metadata = {
    scope: "package",
    title: "Test Rule",
    description: "A test rule",
    custom: {
        category: "Test",
        rulename: "TestRule",
        severity: "LOW",
        rulenumber: "001_0001",
        remediation: "Fix it",
        input: ".*\\.yaml"
    }
};

function rule(input) {
    return { allow: true, errors: [] };
}
`
		err := os.WriteFile(filepath.Join(tempDir, "rule.js"), []byte(jsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write js file: %v", err)
		}

		// Create a test file (should be ignored)
		err = os.WriteFile(filepath.Join(tempDir, "rule_test.js"), []byte(jsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		rules, err := ReadRulesMetadata(tempDir)
		if err != nil {
			t.Fatalf("Failed to read rules metadata: %v", err)
		}

		if len(rules) != 1 {
			t.Errorf("Expected 1 rule (test file should be ignored), got %d", len(rules))
		}
	})
}

func TestEvalAll(t *testing.T) {
	t.Run("all rules pass", func(t *testing.T) {
		err := EvalAll("./../resources/rules", "./../resources/modelsource-v1", "", "", false, false)
		if err != nil {
			t.Errorf("Expected no failures: %v", err)
		}
	})

	t.Run("with xunit report", func(t *testing.T) {
		tempDir := t.TempDir()
		xunitPath := filepath.Join(tempDir, "report.xml")

		err := EvalAll("./../resources/rules", "./../resources/modelsource-v1", xunitPath, "", false, false)
		if err != nil {
			t.Errorf("Expected no failures: %v", err)
		}

		// Verify report was created
		if _, err := os.Stat(xunitPath); os.IsNotExist(err) {
			t.Error("Expected xunit report to be created")
		}
	})

	t.Run("with json report", func(t *testing.T) {
		tempDir := t.TempDir()
		jsonPath := filepath.Join(tempDir, "report.json")

		err := EvalAll("./../resources/rules", "./../resources/modelsource-v1", "", jsonPath, false, false)
		if err != nil {
			t.Errorf("Expected no failures: %v", err)
		}

		// Verify report was created
		if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
			t.Error("Expected json report to be created")
		}
	})
}

func TestEvalAllWithResults(t *testing.T) {
	t.Run("returns results", func(t *testing.T) {
		result, err := EvalAllWithResults("./../resources/rules", "./../resources/modelsource-v1", "", "", false, false)
		if err != nil {
			t.Errorf("Expected no failures: %v", err)
		}

		testSuites, ok := result.(TestSuites)
		if !ok {
			t.Fatal("Expected result to be TestSuites")
		}

		if len(testSuites.Testsuites) == 0 {
			t.Error("Expected at least one testsuite")
		}
		if len(testSuites.Rules) == 0 {
			t.Error("Expected at least one rule in results")
		}
	})

	t.Run("reports failures correctly", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a failing rule with proper metadata structure
		jsContent := `
const metadata = {
    scope: "package",
    title: "Always Fail Rule",
    description: "This rule always fails",
    custom: {
        category: "Test",
        rulename: "AlwaysFailRule",
        severity: "HIGH",
        rulenumber: "099_0099",
        remediation: "Cannot be fixed",
        input: ".*\\.yaml"
    }
};

function rule(input) {
    return { allow: false, errors: ["Always fails"] };
}
`
		jsPath := filepath.Join(tempDir, "fail.js")
		err := os.WriteFile(jsPath, []byte(jsContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write js file: %v", err)
		}

		// Create input file
		yamlPath := filepath.Join(tempDir, "input.yaml")
		err = os.WriteFile(yamlPath, []byte(`Name: "Test"`), 0644)
		if err != nil {
			t.Fatalf("Failed to write yaml file: %v", err)
		}

		_, err = EvalAllWithResults(tempDir, tempDir, "", "", false, false)
		if err == nil {
			t.Error("Expected error due to failures")
		}
	})
}

func TestEvalTestsuite_PatternMatching(t *testing.T) {
	tempDir := t.TempDir()

	// Create a rule with specific pattern
	jsContent := `
const metadata = {
    title: "Pattern Test Rule",
    custom: { rulenumber: "099_0005", input: ".*\\.entity\\.yaml" }
};

function rule(input) {
    return { allow: true, errors: [] };
}
`
	jsPath := filepath.Join(tempDir, "pattern_test.js")
	err := os.WriteFile(jsPath, []byte(jsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write js file: %v", err)
	}

	// Create matching file
	err = os.WriteFile(filepath.Join(tempDir, "test.entity.yaml"), []byte(`Name: "Test"`), 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml file: %v", err)
	}

	// Create non-matching file
	err = os.WriteFile(filepath.Join(tempDir, "test.other.yaml"), []byte(`Name: "Test"`), 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml file: %v", err)
	}

	rule := Rule{
		Path:        jsPath,
		Pattern:     ".*\\.entity\\.yaml",
		PackageName: jsPath,
		RuleNumber:  "099_0005",
		Language:    LanguageJavascript,
	}

	result, err := evalTestsuite(rule, tempDir, false, false)
	if err != nil {
		t.Fatalf("Failed to evaluate testsuite: %v", err)
	}

	// Only the .entity.yaml file should match
	if result.Tests != 1 {
		t.Errorf("Expected 1 test (pattern should match only .entity.yaml), got %d", result.Tests)
	}
}

func TestEvalTestsuite_TimeTracking(t *testing.T) {
	tempDir := t.TempDir()

	jsContent := `
const metadata = {
    title: "Time Test Rule",
    custom: { rulenumber: "099_0006", input: ".*\\.yaml" }
};

function rule(input) {
    return { allow: true, errors: [] };
}
`
	jsPath := filepath.Join(tempDir, "time_test.js")
	err := os.WriteFile(jsPath, []byte(jsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write js file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "input.yaml"), []byte(`Name: "Test"`), 0644)
	if err != nil {
		t.Fatalf("Failed to write yaml file: %v", err)
	}

	rule := Rule{
		Path:        jsPath,
		Pattern:     ".*\\.yaml",
		PackageName: jsPath,
		RuleNumber:  "099_0006",
		Language:    LanguageJavascript,
	}

	result, err := evalTestsuite(rule, tempDir, false, false)
	if err != nil {
		t.Fatalf("Failed to evaluate testsuite: %v", err)
	}

	if result.Time <= 0 {
		t.Error("Expected positive total time")
	}

	for _, tc := range result.Testcases {
		if tc.Time <= 0 {
			t.Errorf("Expected positive time for testcase %s", tc.Name)
		}
	}
}
