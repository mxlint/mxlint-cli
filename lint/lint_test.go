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
	t.Run("single policy passes", func(t *testing.T) {
		result, err := evalTestsuite("./../policies/001_project_settings/001_0003_security_checks.rego", "./../modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result.Failures != 0 {
			t.Errorf("Policy passes")
		}
	})
}

func TestLintBundle(t *testing.T) {
	t.Run("all-policy", func(t *testing.T) {
		err := EvalAll("./../policies", "./../modelsource", "")

		if err == nil {
			t.Errorf("We expect failures in the reference model")
		}
	})
}
