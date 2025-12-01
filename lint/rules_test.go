package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllRules(t *testing.T) {
	t.Run("all-rules", func(t *testing.T) {
		err := TestAll("./../resources/rules")

		if err != nil {
			t.Errorf("No failures expected: %v", err)
		}
	})
}

func TestDuplicateRuleNumbers(t *testing.T) {
	// Create a temporary directory for test rules
	tempDir := t.TempDir()

	// Create two rules with the same rule number
	rule1Content := `# METADATA
# scope: package
# title: Test Rule 1
# description: First test rule
# custom:
#  category: Test
#  rulename: TestRule1
#  severity: LOW
#  rulenumber: 001_0001
#  remediation: Fix it
#  input: .*test\.yaml
package app.test.rule1
import rego.v1

default allow := false
allow if count(errors) == 0
errors contains "test error" if false
`

	rule2Content := `# METADATA
# scope: package
# title: Test Rule 2
# description: Second test rule with same number
# custom:
#  category: Test
#  rulename: TestRule2
#  severity: LOW
#  rulenumber: 001_0001
#  remediation: Fix it
#  input: .*test\.yaml
package app.test.rule2
import rego.v1

default allow := false
allow if count(errors) == 0
errors contains "test error" if false
`

	// Write the rule files
	rule1Path := filepath.Join(tempDir, "rule1.rego")
	rule2Path := filepath.Join(tempDir, "rule2.rego")

	err := os.WriteFile(rule1Path, []byte(rule1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write rule1: %v", err)
	}

	err = os.WriteFile(rule2Path, []byte(rule2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write rule2: %v", err)
	}

	// Test that duplicate rule numbers are detected
	err = TestAll(tempDir)
	if err == nil {
		t.Error("Expected error for duplicate rule numbers, but got nil")
	}

	if err != nil && err.Error() != "found duplicate rule numbers" {
		t.Errorf("Expected 'found duplicate rule numbers' error, got: %v", err)
	}
}

func TestUniqueRuleNumbers(t *testing.T) {
	// Create a temporary directory for test rules
	tempDir := t.TempDir()

	// Create two rules with different rule numbers
	rule1Content := `# METADATA
# scope: package
# title: Test Rule 1
# description: First test rule
# custom:
#  category: Test
#  rulename: TestRule1
#  severity: LOW
#  rulenumber: 001_0001
#  remediation: Fix it
#  input: .*test\.yaml
package app.test.rule1
import rego.v1

default allow := false
allow if count(errors) == 0
errors contains "test error" if false
`

	rule2Content := `# METADATA
# scope: package
# title: Test Rule 2
# description: Second test rule with different number
# custom:
#  category: Test
#  rulename: TestRule2
#  severity: LOW
#  rulenumber: 001_0002
#  remediation: Fix it
#  input: .*test\.yaml
package app.test.rule2
import rego.v1

default allow := false
allow if count(errors) == 0
errors contains "test error" if false
`

	// Write the rule files
	rule1Path := filepath.Join(tempDir, "rule1.rego")
	rule2Path := filepath.Join(tempDir, "rule2.rego")

	// Create test files for the rules
	testFile1 := filepath.Join(tempDir, "rule1_test.yaml")
	testFile2 := filepath.Join(tempDir, "rule2_test.yaml")

	testContent := `TestCases:
  - name: "test case"
    input:
      test: true
    allow: true
`

	err := os.WriteFile(rule1Path, []byte(rule1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write rule1: %v", err)
	}

	err = os.WriteFile(rule2Path, []byte(rule2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write rule2: %v", err)
	}

	err = os.WriteFile(testFile1, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}

	err = os.WriteFile(testFile2, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	// Test that unique rule numbers pass validation
	err = TestAll(tempDir)
	if err != nil {
		// We expect errors from running the actual tests, but not from duplicate rule numbers
		if err.Error() == "found duplicate rule numbers" {
			t.Errorf("Should not get duplicate rule numbers error with unique numbers: %v", err)
		}
		// Other errors from test execution are fine for this test
	}
}
