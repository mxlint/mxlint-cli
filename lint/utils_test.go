package lint

import (
	"testing"
)

func TestParseNoqaDirective(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		expectedSkipAll bool
		expectedRules  []string
		expectedReason string
	}{
		{
			name:           "Skip all rules with #noqa",
			line:           "#noqa",
			expectedSkipAll: true,
			expectedRules:  nil,
			expectedReason: "#noqa",
		},
		{
			name:           "Skip all rules with # noqa",
			line:           "# noqa",
			expectedSkipAll: true,
			expectedRules:  nil,
			expectedReason: "# noqa",
		},
		{
			name:           "Skip all rules with message",
			line:           "#noqa This is a reason",
			expectedSkipAll: true,
			expectedRules:  nil,
			expectedReason: "#noqa This is a reason",
		},
		{
			name:           "Skip specific rule",
			line:           "#noqa:001_0002",
			expectedSkipAll: false,
			expectedRules:  []string{"001_0002"},
			expectedReason: "#noqa:001_0002",
		},
		{
			name:           "Skip multiple rules",
			line:           "#noqa:001_0002,001_0003",
			expectedSkipAll: false,
			expectedRules:  []string{"001_0002", "001_0003"},
			expectedReason: "#noqa:001_0002,001_0003",
		},
		{
			name:           "Skip multiple rules with reason",
			line:           "#noqa:001_0002,001_0003 some reason here",
			expectedSkipAll: false,
			expectedRules:  []string{"001_0002", "001_0003"},
			expectedReason: "#noqa:001_0002,001_0003 some reason here",
		},
		{
			name:           "Case insensitive",
			line:           "#NOQA:001_0002",
			expectedSkipAll: false,
			expectedRules:  []string{"001_0002"},
			expectedReason: "#NOQA:001_0002",
		},
		{
			name:           "Not a noqa directive",
			line:           "This is not a noqa",
			expectedSkipAll: false,
			expectedRules:  nil,
			expectedReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipAll, rules, reason := parseNoqaDirective(tt.line)

			if skipAll != tt.expectedSkipAll {
				t.Errorf("Expected skipAll=%v, got %v", tt.expectedSkipAll, skipAll)
			}

			if len(rules) != len(tt.expectedRules) {
				t.Errorf("Expected %d rules, got %d", len(tt.expectedRules), len(rules))
			} else {
				for i, rule := range rules {
					if rule != tt.expectedRules[i] {
						t.Errorf("Expected rule[%d]=%s, got %s", i, tt.expectedRules[i], rule)
					}
				}
			}

			if reason != tt.expectedReason {
				t.Errorf("Expected reason=%s, got %s", tt.expectedReason, reason)
			}
		})
	}
}

func TestShouldSkipRule(t *testing.T) {
	tests := []struct {
		name           string
		documentation  string
		ruleNumber     string
		expectedSkip   bool
		expectedReason string
	}{
		{
			name:           "Skip all rules",
			documentation:  "#noqa",
			ruleNumber:     "001_0002",
			expectedSkip:   true,
			expectedReason: "#noqa",
		},
		{
			name:           "Skip specific rule - match",
			documentation:  "#noqa:001_0002,001_0003",
			ruleNumber:     "001_0002",
			expectedSkip:   true,
			expectedReason: "#noqa:001_0002,001_0003",
		},
		{
			name:           "Skip specific rule - no match",
			documentation:  "#noqa:001_0002,001_0003",
			ruleNumber:     "001_0004",
			expectedSkip:   false,
			expectedReason: "",
		},
		{
			name:           "Multiple lines with noqa",
			documentation:  "Some text\n#noqa:001_0002 reason\nMore text",
			ruleNumber:     "001_0002",
			expectedSkip:   true,
			expectedReason: "#noqa:001_0002 reason",
		},
		{
			name:           "No noqa directive",
			documentation:  "This is just documentation",
			ruleNumber:     "001_0002",
			expectedSkip:   false,
			expectedReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skip, reason := shouldSkipRule(tt.documentation, tt.ruleNumber, false)

			if skip != tt.expectedSkip {
				t.Errorf("Expected skip=%v, got %v", tt.expectedSkip, skip)
			}

			if reason != tt.expectedReason {
				t.Errorf("Expected reason=%s, got %s", tt.expectedReason, reason)
			}
		})
	}
}

func TestShouldSkipRuleWithIgnoreNoqa(t *testing.T) {
	tests := []struct {
		name          string
		documentation string
		ruleNumber    string
	}{
		{
			name:          "Skip all rules ignored",
			documentation: "#noqa",
			ruleNumber:    "001_0002",
		},
		{
			name:          "Skip specific rule ignored",
			documentation: "#noqa:001_0002,001_0003",
			ruleNumber:    "001_0002",
		},
		{
			name:          "Multiple lines with noqa ignored",
			documentation: "Some text\n#noqa:001_0002 reason\nMore text",
			ruleNumber:    "001_0002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skip, reason := shouldSkipRule(tt.documentation, tt.ruleNumber, true)

			if skip {
				t.Errorf("Expected skip=false when ignoreNoqa=true, got %v", skip)
			}

			if reason != "" {
				t.Errorf("Expected empty reason when ignoreNoqa=true, got %s", reason)
			}
		})
	}
}

