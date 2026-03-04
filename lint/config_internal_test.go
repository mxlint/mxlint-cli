package lint

import (
	"slices"
	"testing"
)

func TestNormalizeSkipPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "trim and dot prefix", input: " ./example/doc ", expected: "example/doc"},
		{name: "leading slash", input: "/example/doc.yaml", expected: "example/doc.yaml"},
		{name: "collapse separators", input: "example//nested/../doc", expected: "example/doc"},
		{name: "dot path becomes empty", input: ".", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSkipPath(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestBuildSkipPathCandidates(t *testing.T) {
	candidates := buildSkipPathCandidates(
		"/tmp/modelsource/example/doc.yaml",
		"/tmp/modelsource",
	)

	expected := []string{
		"tmp/modelsource/example/doc.yaml",
		"tmp/modelsource/example/doc",
		"example/doc.yaml",
		"example/doc",
	}

	for _, want := range expected {
		if !slices.Contains(candidates, want) {
			t.Fatalf("expected candidate %q to be present in %#v", want, candidates)
		}
	}
}

func TestFormatConfigSkipReason(t *testing.T) {
	tests := []struct {
		name     string
		entry    ConfigSkipRule
		expected string
	}{
		{
			name: "reason has priority",
			entry: ConfigSkipRule{
				Reason: "from config",
				Date:   "2026-03-04",
			},
			expected: "from config",
		},
		{
			name: "date fallback",
			entry: ConfigSkipRule{
				Date: "2026-03-04",
			},
			expected: "Skipped by lint.skip config (2026-03-04)",
		},
		{
			name:     "default fallback",
			entry:    ConfigSkipRule{},
			expected: "Skipped by lint.skip config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatConfigSkipReason(tt.entry)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestShouldSkipByConfig(t *testing.T) {
	SetConfig(&Config{
		Lint: ConfigLintSpec{
			Skip: map[string][]ConfigSkipRule{
				"example/doc": []ConfigSkipRule{
					{Rule: "001_002", Reason: "specific skip"},
					{Rule: "*", Reason: "wildcard skip"},
				},
			},
		},
	})
	t.Cleanup(func() {
		SetConfig(&Config{})
	})

	t.Run("specific rule match", func(t *testing.T) {
		skip, reason := shouldSkipByConfig("/tmp/modelsource/example/doc.yaml", "001_002", "/tmp/modelsource")
		if !skip {
			t.Fatal("expected skip=true")
		}
		if reason != "specific skip" {
			t.Fatalf("expected specific reason, got %q", reason)
		}
	})

	t.Run("wildcard fallback", func(t *testing.T) {
		skip, reason := shouldSkipByConfig("/tmp/modelsource/example/doc.yaml", "099_999", "/tmp/modelsource")
		if !skip {
			t.Fatal("expected wildcard skip=true")
		}
		if reason != "wildcard skip" {
			t.Fatalf("expected wildcard reason, got %q", reason)
		}
	})

	t.Run("no path match", func(t *testing.T) {
		skip, reason := shouldSkipByConfig("/tmp/modelsource/example/other.yaml", "001_002", "/tmp/modelsource")
		if skip {
			t.Fatal("expected skip=false")
		}
		if reason != "" {
			t.Fatalf("expected empty reason, got %q", reason)
		}
	})
}
