package lint

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v2"
)

func TestAll(rulesPath string) error {

	allRules, err := ReadRulesMetadata(rulesPath)

	if err != nil {
		return err
	}

	for _, rule := range allRules {
		runTestCases(rule)
	}
	return nil
}

func runTestCases(rule Rule) error {
	log.Infof(">> %s", rule.Path)
	if rule.Language == LanguageJavascript {
		err := runJavaScriptTestCases(rule)
		if err != nil {
			log.Errorf("Failed: %v", err)
		}
	} else if rule.Language == LanguageRego {
		err := runRegoTestCases(rule)
		if err != nil {
			log.Errorf("Failed: %v", err)
		}
	} else {
		log.Warnf("Skipped unsupported rule %s.", rule.Path)
	}
	return nil
}

func runJavaScriptTestCases(rule Rule) error {

	ruleContent, err := os.ReadFile(rule.Path)
	if err != nil {
		return err
	}

	testFilePath := strings.Replace(rule.Path, ".js", "_test.yaml", 1)
	testCases, err := readTestCases(testFilePath)
	if err != nil {
		return err
	}

	for _, testCase := range testCases {
		testCaseMap := testCase.(map[interface{}]interface{})
		input := convertToStringKeyMap(testCaseMap["input"].(map[interface{}]interface{}))
		allow := testCaseMap["allow"].(bool)

		vm := goja.New()
		_, err = vm.RunString(string(ruleContent))
		if err != nil {
			panic(err)
		}

		ruleFunction, ok := goja.AssertFunction(vm.Get("rule"))
		if !ok {
			panic("rule(...) function not found")
		}

		res, err := ruleFunction(goja.Undefined(), vm.ToValue(input))
		if err != nil {
			panic(err)
		}

		rs := res.Export().(map[string]interface{})

		result := rs["allow"].(bool)
		errors := rs["errors"].([]interface{})

		if result != allow {
			for _, error := range errors {
				log.Errorf("Error: %s", error)
			}
			return fmt.Errorf("FAIL %s: Expected %v, got: %v", testCaseMap["name"], allow, result)
		} else {
			log.Infof("PASS  %s ", testCaseMap["name"])
		}
	}

	return nil
}

func runRegoTestCases(rule Rule) error {

	packageName := getPackageName(rule.Path)
	queryString := fmt.Sprintf("data.%s.allow", packageName)
	testFilePath := strings.Replace(rule.Path, ".rego", "_test.yaml", 1)

	testCases, err := readTestCases(testFilePath)
	if err != nil {
		return err
	}

	for _, testCase := range testCases {
		testCaseMap := testCase.(map[interface{}]interface{})
		input := convertToStringKeyMap(testCaseMap["input"].(map[interface{}]interface{}))
		allow := testCaseMap["allow"].(bool)

		ctx := context.Background()

		r := rego.New(
			rego.Query(queryString),
			rego.Load([]string{rule.Path}, nil),
			rego.Input(input),
			rego.Trace(true),
		)

		rs, err := r.Eval(ctx)
		if err != nil {
			log.Fatal(err)
			return err
		}

		log.Debugf("Result: %v", rs)

		result := rs[0].Expressions[0].Value.(bool)
		if result != allow {
			log.Errorf("FAIL %s: Expected: %v, got: %v", testCaseMap["name"], allow, result)
			errors := rs[0].Expressions[1].Value.([]interface{})
			for _, error := range errors {
				log.Errorf("Error: %s", error)
			}
		} else {
			log.Infof("PASS  %s", testCaseMap["name"])
		}
	}

	return nil
}

// convertToStringKeyMap converts a map[interface{}]interface{} to map[string]interface{}
func convertToStringKeyMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key := fmt.Sprintf("%v", k)
		switch val := v.(type) {
		case map[interface{}]interface{}:
			result[key] = convertToStringKeyMap(val)
		case []interface{}:
			result[key] = convertSlice(val)
		default:
			result[key] = v
		}
	}
	return result
}

// convertSlice converts a slice of interface{} to a slice of properly converted values
func convertSlice(s []interface{}) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		switch val := v.(type) {
		case map[interface{}]interface{}:
			result[i] = convertToStringKeyMap(val)
		case []interface{}:
			result[i] = convertSlice(val)
		default:
			result[i] = v
		}
	}
	return result
}

func getPackageName(rulePath string) string {
	fileContent, err := os.ReadFile(rulePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(fileContent), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "package ") {
			return strings.TrimSpace(strings.Split(line, " ")[1])
		}
	}
	return ""
}

func readTestCases(testFilePath string) ([]interface{}, error) {

	testFileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = yaml.Unmarshal(testFileContent, &data)
	if err != nil {
		log.Errorf("Failed to parse test file %s: %v", testFilePath, err)
		return nil, err
	}
	testCases := data["TestCases"].([]interface{})

	return testCases, nil
}
