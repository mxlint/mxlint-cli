package lint

import (
	"os"
	"path/filepath"
	"strings"
)

func readPoliciesMetadata(policiesPath string) ([]Policy, error) {
	policies := make([]Policy, 0)
	filepath.Walk(policiesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(info.Name(), "_test.rego") && strings.HasSuffix(info.Name(), ".rego") {
			policy, err := parsePolicyMetadata(path)
			if err != nil {
				return err
			}
			policies = append(policies, *policy)
		}
		return nil
	})
	return policies, nil
}

func parsePolicyMetadata(policyPath string) (*Policy, error) {

	log.Debugf("reading policy %s", policyPath)

	// read the policy file
	policyFile, err := os.Open(policyPath)
	if err != nil {
		return nil, err
	}
	defer policyFile.Close()

	policyContent, err := os.ReadFile(policyPath)
	if err != nil {
		return nil, err
	}

	var packageName string = ""
	var pattern string = ""
	var skipReason string = ""
	var title string = ""
	var description string = ""
	var category string = ""
	var severity string = ""
	var ruleNumber string = ""
	var remediation string = ""
	var ruleName string = ""
	var key string = ""
	var value string = ""

	lines := strings.Split(string(policyContent), "\n")

	for _, line := range lines {
		tokens := strings.Split(line, "package ")
		if len(tokens) > 1 && packageName == "" {
			packageName = tokens[1]
		}
		// only read the comments as that is where the metadata is stored
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		// strip the comment prefix
		line = strings.TrimPrefix(line, "# ")
		tokens = strings.SplitN(line, ":", 2)
		if len(tokens) == 2 {
			key = strings.Trim(strings.TrimSpace(tokens[0]), "\"")
			value = strings.Trim(strings.TrimSpace(tokens[1]), "\"")
		}
		switch key {
		case "input":
			pattern = value
		case "skip":
			skipReason = value
		case "title":
			title = value
		case "description":
			description = value
		case "category":
			category = value
		case "rulename":
			ruleName = value
		case "severity":
			severity = value
		case "rulenumber":
			ruleNumber = value
		case "remediation":
			remediation = value
		}
	}

	policy := &Policy{
		Title:       title,
		Description: description,
		Category:    category,
		Severity:    severity,
		RuleNumber:  ruleNumber,
		Remediation: remediation,
		RuleName:    ruleName,
		Path:        policyPath,
		SkipReason:  skipReason,
		Pattern:     pattern,
		PackageName: packageName,
	}
	return policy, nil
}
