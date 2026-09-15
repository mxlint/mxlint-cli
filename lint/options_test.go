package lint

import "testing"

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func TestEffectiveLintConcurrency_DefaultIsBounded(t *testing.T) {
	SetConfig(&Config{})
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	value := effectiveLintConcurrency(100)
	if value < 1 || value > defaultMaxLintConcurrency {
		t.Fatalf("expected default concurrency within [1,%d], got %d", defaultMaxLintConcurrency, value)
	}
}

func TestEffectiveLintConcurrency_UsesConfigWhenProvided(t *testing.T) {
	SetConfig(&Config{
		Lint: ConfigLintSpec{
			Concurrency: intPtr(2),
		},
	})
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	value := effectiveLintConcurrency(10)
	if value != 2 {
		t.Fatalf("expected configured concurrency 2, got %d", value)
	}
}

func TestEffectiveLintConcurrency_CapsToRuleCount(t *testing.T) {
	SetConfig(&Config{
		Lint: ConfigLintSpec{
			Concurrency: intPtr(8),
		},
	})
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	value := effectiveLintConcurrency(3)
	if value != 3 {
		t.Fatalf("expected concurrency capped to rule count 3, got %d", value)
	}
}

func TestRegoTraceEnabled(t *testing.T) {
	SetConfig(&Config{
		Lint: ConfigLintSpec{
			RegoTrace: boolPtr(true),
		},
	})
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	if !regoTraceEnabled() {
		t.Fatal("expected regoTraceEnabled to return true")
	}
}
