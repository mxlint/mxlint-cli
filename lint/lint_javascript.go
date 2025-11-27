package lint

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grafana/sobek"
	"gopkg.in/yaml.v3"
)

func evalTestcase_Javascript(rulePath string, inputFilePath string, ruleNumber string) (*Testcase, error) {
	ruleContent, _ := os.ReadFile(rulePath)
	log.Debugf("js file: \n%s", ruleContent)

	documentContent, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Errorf("Error reading YAML file: %s\n", err)
		return nil, err
	}

	// parse the input file as YAML
	var data map[string]interface{}
	var node yaml.Node
	err = yaml.Unmarshal(documentContent, &node)
	if err != nil {
		log.Errorf("Error parsing YAML file: %s\n", err)
		return nil, err
	}
	err = node.Decode(&data)
	if err != nil {
		log.Errorf("Error decoding YAML file: %s\n", err)
		return nil, err
	}

	// Check if this rule should be skipped based on noqa directives
	if doc, ok := data["Documentation"].(string); ok {
		shouldSkip, reason := shouldSkipRule(doc, ruleNumber)
		if shouldSkip {
			return &Testcase{
				Name:    inputFilePath,
				Time:    0,
				Skipped: &Skipped{Message: reason},
			}, nil
		}
	}

	startTime := time.Now()

	vm := sobek.New()
	_, err = vm.RunString(string(ruleContent))
	if err != nil {
		panic(err)
	}

	ruleFunction, ok := sobek.AssertFunction(vm.Get("rule"))
	if !ok {
		panic("rule(...) function not found")
	}

	res, err := ruleFunction(sobek.Undefined(), vm.ToValue(data))
	if err != nil {
		panic(err)
	}

	rs := res.Export().(map[string]interface{})

	duration := time.Since(startTime)

	var failure *Failure = nil

	log.Debugf("Result: %v", rs)
	result := rs["allow"].(bool)
	errors := rs["errors"].([]interface{})
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

func parseRuleMetadata_Javascript(rulePath string) (*Rule, error) {

	log.Debugf("reading rule %s", rulePath)

	// read the rule file
	ruleContent, err := os.ReadFile(rulePath)
	if err != nil {
		return nil, err
	}

	// use sobek to extract the metadata from the rule
	vm := sobek.New()
	_, err = vm.RunString(string(ruleContent))
	if err != nil {
		panic(err)
	}
	// FIXME: handle the case where metadata is not defined correctly
	metadata := vm.Get("metadata")
	metadataMap := metadata.ToObject(vm)

	var packageName string = rulePath
	var title string = metadataMap.Get("title").String()
	var description string = metadataMap.Get("description").String()

	// custom metadata
	custom := metadataMap.Get("custom").ToObject(vm)
	var category string = custom.Get("category").String()
	var severity string = custom.Get("severity").String()
	var ruleNumber string = custom.Get("rulenumber").String()
	var remediation string = custom.Get("remediation").String()
	var ruleName string = custom.Get("rulename").String()
	var pattern string = custom.Get("input").String()

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
		Language:    LanguageJavascript,
	}
	return rule, nil
}
