package lint

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
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
	var node yaml.Node
	err = yaml.Unmarshal(yamlFile, &node)
	if err != nil {
		log.Errorf("Error parsing YAML file: %s\n", err)
		return nil, err
	}
	err = node.Decode(&data)
	if err != nil {
		log.Errorf("Error decoding YAML file: %s\n", err)
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
	var title string = ""
	var description string = ""
	var category string = ""
	var severity string = ""
	var ruleNumber string = ""
	var remediation string = ""
	var ruleName string = ""

	lines := strings.Split(string(ruleContent), "\n")

	// extract package name and collect metadata block
	var metadataLines []string
	inMetadata := false

	for _, line := range lines {
		tokens := strings.Split(line, "package ")
		if len(tokens) > 1 && packageName == "" {
			packageName = tokens[1]
		}

		// look for OPA metadata marker
		if strings.TrimSpace(line) == "# METADATA" {
			inMetadata = true
			continue
		}

		// collect metadata comment lines
		if inMetadata {
			if strings.HasPrefix(line, "# ") {
				metadataLines = append(metadataLines, strings.TrimPrefix(line, "# "))
			} else if strings.HasPrefix(line, "#") && strings.TrimSpace(line) == "#" {
				metadataLines = append(metadataLines, "")
			} else {
				break
			}
		}
	}

	// parse metadata as YAML
	if len(metadataLines) > 0 {
		yamlContent := strings.Join(metadataLines, "\n")
		log.Debugf("Parsing metadata YAML:\n%s", yamlContent)

		var metadata struct {
			Title       string `yaml:"title"`
			Description string `yaml:"description"`
			Custom      struct {
				Category    string `yaml:"category"`
				RuleName    string `yaml:"rulename"`
				Severity    string `yaml:"severity"`
				RuleNumber  string `yaml:"rulenumber"`
				Remediation string `yaml:"remediation"`
				Input       string `yaml:"input"`
			} `yaml:"custom"`
		}

		err = yaml.Unmarshal([]byte(yamlContent), &metadata)
		if err != nil {
			log.Warnf("Error parsing metadata YAML: %s", err)
			// continue with empty metadata on parse failure
		} else {
			title = metadata.Title
			description = metadata.Description
			category = metadata.Custom.Category
			ruleName = metadata.Custom.RuleName
			severity = metadata.Custom.Severity
			ruleNumber = metadata.Custom.RuleNumber
			remediation = metadata.Custom.Remediation
			pattern = metadata.Custom.Input
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
		Pattern:     pattern,
		PackageName: packageName,
		Language:    LanguageRego,
	}
	return rule, nil
}
