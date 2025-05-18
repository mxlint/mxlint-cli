package lint

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/ghodss/yaml"
)

func evalTestcase_Javascript(rulePath string, inputFilePath string) (*Testcase, error) {
	ruleContent, _ := os.ReadFile(rulePath)
	log.Debugf("js file: \n%s", ruleContent)

	documentContent, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Errorf("Error reading YAML file: %s\n", err)
		return nil, err
	}

	// parse the input file as YAML
	var data map[string]interface{}
	err = yaml.Unmarshal(documentContent, &data)
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

	startTime := time.Now()

	vm := goja.New()
	_, err = vm.RunString(string(ruleContent))
	if err != nil {
		panic(err)
	}

	ruleFunction, ok := goja.AssertFunction(vm.Get("rule"))
	if !ok {
		panic("rule(...) function not found")
	}

	res, err := ruleFunction(goja.Undefined(), vm.ToValue(data))
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

	// use goja to extract the metadata from the rule
	vm := goja.New()
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
	var skipReason string = ""
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
		SkipReason:  skipReason,
		Pattern:     pattern,
		PackageName: packageName,
		Language:    LanguageJavascript,
	}
	return rule, nil
}
