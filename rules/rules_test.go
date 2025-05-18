package rules

import (
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
