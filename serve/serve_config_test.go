package serve

import (
	"testing"

	"github.com/mxlint/mxlint-cli/lint"
)

func TestEffectiveLintUseCacheForServe(t *testing.T) {
	falseValue := false

	tests := []struct {
		name     string
		config   *lint.Config
		expected bool
	}{
		{
			name:     "nil config defaults to cache enabled",
			config:   nil,
			expected: true,
		},
		{
			name: "cache enabled by default",
			config: &lint.Config{
				Cache: lint.ConfigCacheSpec{
					Enable: nil,
				},
			},
			expected: true,
		},
		{
			name: "cache disabled by cache.enable",
			config: &lint.Config{
				Cache: lint.ConfigCacheSpec{
					Enable: &falseValue,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := effectiveLintUseCacheForServe(tt.config); got != tt.expected {
				t.Fatalf("expected %t, got %t", tt.expected, got)
			}
		})
	}
}
