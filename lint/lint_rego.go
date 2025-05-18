package lint

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/open-policy-agent/opa/rego"
)

func evalTestcase_Rego(rulePath string, queryString string, inputFilePath string) (*Testcase, error) {
	regoFile, _ := os.ReadFile(rulePath)
	log.Debugf("rego file: \n%s", regoFile)

	yamlFile, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Errorf("Error reading YAML file: %s\n", err)
		return nil, err
	}

	var data map[string]interface{}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		log.Errorf("Error parsing YAML file: %s\n", err)
		return nil, err
	}

	// if data["Documentation"] contains #noqa, skip the testcase; Documentation attribute might not exist
	if doc, ok := data["Documentation"].(string); ok {
		lines := strings.Split(doc, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			lineLower := strings.ToLower(line)
			if strings.HasPrefix(lineLower, NOQA) || strings.HasPrefix(lineLower, NOQA_ALIAS) {
				return &Testcase{
					Name:    inputFilePath,
					Time:    0,
					Skipped: &Skipped{Message: line},
				}, nil
			}
		}
	}

	ctx := context.Background()

	startTime := time.Now()
	r := rego.New(
		rego.Query(queryString),
		rego.Load([]string{rulePath}, nil),
		rego.Input(data),
		rego.Trace(true),
	)

	rs, err := r.Eval(ctx)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	duration := time.Since(startTime)

	var failure *Failure = nil

	log.Debugf("Result: %v", rs)
	rsmap := rs[0].Expressions[0].Value.(map[string]interface{})
	result := rsmap["allow"].(bool)
	errors := rsmap["errors"].([]interface{})
	if !result {
		myErrors := make([]string, 0)
		for _, err := range errors {
			//log.Warnf("Rule failed: %s", err)
			myErrors = append(myErrors, fmt.Sprintf("%s", err))
		}
		failure = &Failure{
			Message: strings.Join(myErrors, "\n"),
			Type:    "AssertionError",
		}
	}
	testcase := &Testcase{
		Name:    inputFilePath,
		Time:    float64(duration.Nanoseconds()) / 1e9, // convert to seconds
		Failure: failure,
		Skipped: nil,
	}
	return testcase, nil
}

func parseRuleMetadata_Rego(rulePath string) (*Rule, error) {

	log.Debugf("reading rule %s", rulePath)

	// read the rule file
	ruleContent, err := os.ReadFile(rulePath)
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

	lines := strings.Split(string(ruleContent), "\n")

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

	rule := &Rule{
		Title:       title,
		Description: description,
		Category:    category,
		Severity:    severity,
		RuleNumber:  ruleNumber,
		Remediation: remediation,
		RuleName:    ruleName,
		Path:        rulePath,
		SkipReason:  skipReason,
		Pattern:     pattern,
		PackageName: packageName,
		Language:    LanguageRego,
	}
	return rule, nil
}
