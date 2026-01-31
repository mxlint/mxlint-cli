package lint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/grafana/sobek"
	"gopkg.in/yaml.v3"
)

type typescriptRuleCacheEntry struct {
	hash string
	code string
}

var typescriptRuleCache = struct {
	mu      sync.RWMutex
	entries map[string]typescriptRuleCacheEntry
}{
	entries: make(map[string]typescriptRuleCacheEntry),
}

func hashRuleContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func formatEsbuildErrors(errors []api.Message) string {
	messages := api.FormatMessages(errors, api.FormatMessagesOptions{
		Kind:  api.ErrorMessage,
		Color: false,
	})
	return strings.Join(messages, "\n")
}

func transpileTypescriptRule(rulePath string) (string, error) {
	content, err := os.ReadFile(rulePath)
	if err != nil {
		return "", err
	}

	ruleHash := hashRuleContent(content)

	typescriptRuleCache.mu.RLock()
	cached, found := typescriptRuleCache.entries[rulePath]
	typescriptRuleCache.mu.RUnlock()
	if found && cached.hash == ruleHash {
		return cached.code, nil
	}

	result := api.Transform(string(content), api.TransformOptions{
		Loader:     api.LoaderTS,
		Target:     api.ES2019,
		Sourcefile: rulePath,
	})
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("failed to transpile typescript rule %s: %s", rulePath, formatEsbuildErrors(result.Errors))
	}

	code := string(result.Code)

	typescriptRuleCache.mu.Lock()
	typescriptRuleCache.entries[rulePath] = typescriptRuleCacheEntry{
		hash: ruleHash,
		code: code,
	}
	typescriptRuleCache.mu.Unlock()

	return code, nil
}

func evalTestcase_Typescript(rulePath string, inputFilePath string, ruleNumber string, ignoreNoqa bool, modelSourcePath string) (*Testcase, error) {
	ruleContent, err := transpileTypescriptRule(rulePath)
	if err != nil {
		return nil, err
	}
	log.Debugf("ts file transpiled: \n%s", ruleContent)

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
		shouldSkip, reason := shouldSkipRule(doc, ruleNumber, ignoreNoqa)
		if shouldSkip {
			return &Testcase{
				Name:    inputFilePath,
				Time:    0,
				Skipped: &Skipped{Message: reason},
			}, nil
		}
	}

	startTime := time.Now()

	// Use the modelsource path as the working directory, falling back to input file's directory
	workingDirectory := modelSourcePath
	if workingDirectory == "" {
		workingDirectory = filepath.Dir(inputFilePath)
	}
	allowedRoot := resolveAllowedRoot(modelSourcePath)
	vm := setupJavascriptVM(workingDirectory, allowedRoot)
	_, err = vm.RunString(ruleContent)
	if err != nil {
		panic(err)
	}

	ruleFunction, ok := sobek.AssertFunction(vm.Get("rule"))
	if !ok {
		panic("rule(...) function not found in rule file: " + rulePath)
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

func parseRuleMetadata_Typescript(rulePath string) (*Rule, error) {

	log.Debugf("reading rule %s", rulePath)

	ruleContent, err := transpileTypescriptRule(rulePath)
	if err != nil {
		return nil, err
	}

	vm := sobek.New()
	_, err = vm.RunString(ruleContent)
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
		Language:    LanguageTypescript,
	}
	return rule, nil
}

func runTypescriptTestCases(rule Rule) error {
	ruleContent, err := transpileTypescriptRule(rule.Path)
	if err != nil {
		return err
	}

	testFilePath := strings.Replace(rule.Path, ".ts", "_test.yaml", 1)
	testCases, err := readTestCases(testFilePath)
	if err != nil {
		return err
	}

	for _, testCase := range testCases {
		var input map[string]interface{}
		var allow bool

		// Handle different map types based on YAML parser
		switch tcMap := testCase.(type) {
		case map[interface{}]interface{}:
			// For yaml.v2
			input = convertToStringKeyMap(tcMap["input"].(map[interface{}]interface{}))
			allow = tcMap["allow"].(bool)
		case map[string]interface{}:
			// For yaml.v3
			inputVal := tcMap["input"]
			switch inputMap := inputVal.(type) {
			case map[interface{}]interface{}:
				input = convertToStringKeyMap(inputMap)
			case map[string]interface{}:
				input = inputMap
			default:
				return fmt.Errorf("unexpected input type: %T", inputVal)
			}
			allow = tcMap["allow"].(bool)
		default:
			return fmt.Errorf("unexpected testCase type: %T", testCase)
		}

		// Use the directory containing the rule file as the working directory
		workingDirectory := filepath.Dir(rule.Path)
		vm := setupJavascriptVM(workingDirectory, workingDirectory)
		_, err = vm.RunString(ruleContent)
		if err != nil {
			panic(err)
		}

		ruleFunction, ok := sobek.AssertFunction(vm.Get("rule"))
		if !ok {
			panic("rule(...) function not found")
		}

		res, err := ruleFunction(sobek.Undefined(), vm.ToValue(input))
		if err != nil {
			panic(err)
		}

		rs := res.Export().(map[string]interface{})

		result := rs["allow"].(bool)
		errors := rs["errors"].([]interface{})

		// Get the test case name
		var name string
		switch tcMap := testCase.(type) {
		case map[interface{}]interface{}:
			if n, ok := tcMap["name"].(string); ok {
				name = n
			} else {
				name = "unnamed test"
			}
		case map[string]interface{}:
			if n, ok := tcMap["name"].(string); ok {
				name = n
			} else {
				name = "unnamed test"
			}
		}

		if result != allow {
			for _, error := range errors {
				log.Errorf("Error: %s", error)
			}
			return fmt.Errorf("FAIL %s: Expected %v, got: %v", name, allow, result)
		} else {
			log.Infof("PASS  %s ", name)
		}
	}

	return nil
}
