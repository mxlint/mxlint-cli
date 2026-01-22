package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTranspileTypescriptRuleUsesCacheWhenUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	rulePath := filepath.Join(tmpDir, "rule.ts")
	content := []byte(`
const metadata = {
  title: "Test",
  description: "Test",
  custom: {
    category: "Test",
    rulename: "TestRule",
    severity: "LOW",
    rulenumber: "000_0001",
    remediation: "None",
    input: ".*"
  }
};

function rule(input: Record<string, unknown> = {}) {
  return { allow: true, errors: [] };
}
`)

	if err := os.WriteFile(rulePath, content, 0o600); err != nil {
		t.Fatalf("failed to write rule: %v", err)
	}

	hash := hashRuleContent(content)
	typescriptRuleCache.mu.Lock()
	typescriptRuleCache.entries[rulePath] = typescriptRuleCacheEntry{
		hash: hash,
		code: "cached",
	}
	typescriptRuleCache.mu.Unlock()

	code, err := transpileTypescriptRule(rulePath)
	if err != nil {
		t.Fatalf("transpile failed: %v", err)
	}

	if code != "cached" {
		t.Fatalf("expected cached output, got %q", code)
	}
}

func TestTranspileTypescriptRuleRefreshesCacheOnChange(t *testing.T) {
	tmpDir := t.TempDir()
	rulePath := filepath.Join(tmpDir, "rule.ts")
	content := []byte(`const metadata = { title: "Test", description: "Test", custom: { category: "Test", rulename: "TestRule", severity: "LOW", rulenumber: "000_0001", remediation: "None", input: ".*" } };
function rule(input: Record<string, unknown> = {}) { return { allow: true, errors: [] }; }`)

	if err := os.WriteFile(rulePath, content, 0o600); err != nil {
		t.Fatalf("failed to write rule: %v", err)
	}

	oldHash := hashRuleContent(content)
	typescriptRuleCache.mu.Lock()
	typescriptRuleCache.entries[rulePath] = typescriptRuleCacheEntry{
		hash: oldHash,
		code: "cached",
	}
	typescriptRuleCache.mu.Unlock()

	updated := []byte(`const metadata = { title: "Test", description: "Updated", custom: { category: "Test", rulename: "TestRule", severity: "LOW", rulenumber: "000_0001", remediation: "None", input: ".*" } };
function rule(input: Record<string, unknown> = {}) { return { allow: false, errors: ["fail"] }; }`)
	if err := os.WriteFile(rulePath, updated, 0o600); err != nil {
		t.Fatalf("failed to update rule: %v", err)
	}

	code, err := transpileTypescriptRule(rulePath)
	if err != nil {
		t.Fatalf("transpile failed: %v", err)
	}
	if code == "cached" {
		t.Fatalf("expected transpiled output, got cached result")
	}

	newHash := hashRuleContent(updated)
	typescriptRuleCache.mu.RLock()
	cached, found := typescriptRuleCache.entries[rulePath]
	typescriptRuleCache.mu.RUnlock()
	if !found {
		t.Fatalf("expected cache entry to be updated")
	}
	if cached.hash != newHash {
		t.Fatalf("expected cache hash %q, got %q", newHash, cached.hash)
	}
}
