package lint

import (
	"testing"
)

// TestAdd tests the Add function to ensure it returns correct results.
func TestLintSingle(t *testing.T) {
	t.Run("single policy skipped", func(t *testing.T) {
		result, err := evalTestsuite("./../policies/001-projectsettings/001-0004-strong_password.rego", "./../modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result.Skipped != 1 {
			t.Errorf("Policy not skipped")
		}
	})
	t.Run("single policy passes", func(t *testing.T) {
		result, err := evalTestsuite("./../policies/001-projectsettings/001-0003-security_enabled.rego", "./../modelsource")

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

		if err != nil {
			t.Errorf("Failed to evaluate: %v", err)
		}
	})
}
