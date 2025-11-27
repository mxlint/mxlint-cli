package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New() // Initialize with a default logger

// SetLogger allows the main application to set the logger, including its configuration.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// parseNoqaDirective parses a noqa directive and returns the list of rules to skip
// and the reason (if provided).
// Supports two formats:
// - "#noqa" or "# noqa" - skips all rules
// - "#noqa:rule1,rule2" or "# noqa:rule1,rule2 reason" - skips specific rules
// Returns: (skipAllRules bool, skipRules []string, reason string)
func parseNoqaDirective(line string) (bool, []string, string) {
	line = strings.TrimSpace(line)
	lineLower := strings.ToLower(line)
	
	// Check if line starts with #noqa or # noqa
	if !strings.HasPrefix(lineLower, NOQA) && !strings.HasPrefix(lineLower, NOQA_ALIAS) {
		return false, nil, ""
	}
	
	// Remove the prefix to get the rest
	var rest string
	if strings.HasPrefix(lineLower, NOQA) {
		rest = strings.TrimSpace(line[len(NOQA):])
	} else {
		rest = strings.TrimSpace(line[len(NOQA_ALIAS):])
	}
	
	// If nothing follows, skip all rules
	if rest == "" {
		return true, nil, line
	}
	
	// Check if it starts with colon (rule-specific noqa)
	if strings.HasPrefix(rest, ":") {
		rest = strings.TrimPrefix(rest, ":")
		
		// Split by space to separate rules from reason
		parts := strings.SplitN(rest, " ", 2)
		rulesStr := strings.TrimSpace(parts[0])
		reason := line // Use full line as reason
		
		// Split rules by comma
		rules := strings.Split(rulesStr, ",")
		skipRules := make([]string, 0, len(rules))
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule != "" {
				skipRules = append(skipRules, rule)
			}
		}
		
		if len(skipRules) > 0 {
			return false, skipRules, reason
		}
	}
	
	// Default: skip all rules with the line as reason
	return true, nil, line
}

// shouldSkipRule checks if a specific rule should be skipped based on noqa directives
// in the documentation field
func shouldSkipRule(documentation string, ruleNumber string, ignoreNoqa bool) (bool, string) {
	// If ignoreNoqa is true, never skip rules based on noqa directives
	if ignoreNoqa {
		return false, ""
	}
	
	if documentation == "" {
		return false, ""
	}
	
	lines := strings.Split(documentation, "\n")
	for _, line := range lines {
		skipAll, skipRules, reason := parseNoqaDirective(line)
		
		// If skipAll is true, skip this rule
		if skipAll {
			return true, reason
		}
		
		// Check if this specific rule is in the skip list
		for _, skipRule := range skipRules {
			if skipRule == ruleNumber {
				return true, reason
			}
		}
	}
	
	return false, ""
}

func expandPaths(pattern string, workingDirectory string) ([]string, error) {
	// backwards compatible with old filepath.glob(...)
	if !strings.HasPrefix(pattern, ".*") {
		oldPattern := pattern
		pattern = strings.ReplaceAll(pattern, "$", "\\$")
		pattern = strings.ReplaceAll(pattern, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "**", ".*")
		log.Infof("Expanded old pattern: %v -> %v", oldPattern, pattern)
	}
	// First get all files recursively under working directory
	var matches []string
	err := filepath.Walk(workingDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}
		// Get relative path from working directory
		relPath, err := filepath.Rel(workingDirectory, path)
		if err != nil {
			return err
		}
		// Check if path matches pattern
		matched, err := regexp.MatchString(pattern, relPath)
		if err != nil {
			log.Errorf("Error matching path %v against pattern %v: %v", relPath, pattern, err)
			return err
		}
		if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		log.Warnf("No matches found for pattern %v ", pattern)
	}
	return matches, nil
}
