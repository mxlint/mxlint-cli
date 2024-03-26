package lint

import (
	"testing"
)

// TestAdd tests the Add function to ensure it returns correct results.
func TestLintSingle(t *testing.T) {
	t.Run("single policy fails", func(t *testing.T) {
		result, err := Eval("./../policies/security/strong_password.rego", "./../modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result != PolicyResultFail {
			t.Errorf("Failed policy")
		}
	})
	t.Run("single policy passes", func(t *testing.T) {
		result, err := Eval("./../policies/security/security_enabled.rego", "./../modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result != PolicyResultPass {
			t.Errorf("Failed policy")
		}
	})
}

func TestLintBundle(t *testing.T) {
	t.Run("all-policy", func(t *testing.T) {
		result, err := EvalAll("./../policies", "./../modelsource")

		if err != nil {
			t.Errorf("Failed to evaluate")
		}

		if result != PolicyResultPass {
			t.Errorf("Failed policy")
		}
	})
}
