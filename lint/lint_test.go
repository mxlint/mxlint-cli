package lint

import (
	"testing"
)

// TestAdd tests the Add function to ensure it returns correct results.
func TestLintSingle(t *testing.T) {
	// t.Run("single policy skipped", func(t *testing.T) {
	// 	result, err := evalTestsuite("./../policies/001_project_settings/001_0004_strong_password.rego", "./../resources/modelsource")

	// 	if err != nil {
	// 		t.Errorf("Failed to evaluate")
	// 	}

	// 	if result.Skipped != 1 {
	// 		t.Errorf("Policy not skipped")
	// 	}
	// })
	t.Run("single Rego rule passes", func(t *testing.T) {
		rule, _ := parseRuleMetadata_Rego("./../resources/rules/001_0003_security_checks.rego")
		result, err := evalTestsuite(*rule, "./../resources/modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result.Failures != 0 {
			t.Errorf("Policy passes")
		}
	})
	t.Run("single JS rule passes", func(t *testing.T) {
		rule, _ := parseRuleMetadata_Javascript("./../resources/rules/001_0002_demo_users_disabled.js")
		result, err := evalTestsuite(*rule, "./../resources/modelsource")

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
		err := EvalAll("./../resources/rules", "./../resources/modelsource", "", "")

		if err != nil {
			t.Errorf("No failures expected: %v", err)
		}
	})
}
