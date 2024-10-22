package lint

import (
	"testing"
)

// TestAdd tests the Add function to ensure it returns correct results.
func TestLintSingle(t *testing.T) {
	// t.Run("single policy skipped", func(t *testing.T) {
	// 	result, err := evalTestsuite("./../policies/001_project_settings/001_0004_strong_password.rego", "./../modelsource")

	// 	if err != nil {
	// 		t.Errorf("Failed to evaluate")
	// 	}

	// 	if result.Skipped != 1 {
	// 		t.Errorf("Policy not skipped")
	// 	}
	// })
	t.Run("single rule passes", func(t *testing.T) {
		rule, _ := parseRuleMetadata("./../resources/rules/001_0003_security_checks.rego")
		result, err := evalTestsuite(*rule, "./../modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result.Failures != 0 {
			t.Errorf("Policy passes")
		}
	})
}

func TestLintBundle(t *testing.T) {
	t.Run("all-rules", func(t *testing.T) {
		err := EvalAll("./../resources/rules", "./../modelsource", "", "")

		if err != nil {
			t.Errorf("No failures expected: %v", err)
		}
	})
}
