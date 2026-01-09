package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grafana/sobek"
	"gopkg.in/yaml.v3"
)

// resolvePath resolves the given path relative to the working directory and validates
// that it stays within the working directory. Returns the absolute path or an error.
func resolvePath(pathArg string, workingDirectory string) (string, error) {
	// Resolve the path relative to working directory
	var fullPath string
	if filepath.IsAbs(pathArg) {
		fullPath = pathArg
	} else {
		fullPath = filepath.Join(workingDirectory, pathArg)
	}

	// Convert both paths to absolute and clean them to resolve any ".." or "." components
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	absFullPath = filepath.Clean(absFullPath)

	absWorkingDir, err := filepath.Abs(workingDirectory)
	if err != nil {
		return "", fmt.Errorf("failed to resolve working directory: %w", err)
	}
	absWorkingDir = filepath.Clean(absWorkingDir)

	// Check that the resolved path is within the working directory
	if !strings.HasPrefix(absFullPath, absWorkingDir+string(filepath.Separator)) && absFullPath != absWorkingDir {
		return "", fmt.Errorf("path %q is outside working directory %q", pathArg, workingDirectory)
	}

	return absFullPath, nil
}

// setupJavascriptVM creates a new sobek VM with the mxlint object exposed.
// The mxlint object provides utility functions for JavaScript rules:
//   - mxlint.readfile(path): Reads a file and returns its contents as a string.
//     The path is resolved relative to the workingDirectory.
//   - mxlint.listdir(path): Lists the contents of a directory and returns an array of filenames.
//     The path is resolved relative to the workingDirectory.
//   - mxlint.isdir(path): Returns true if the path is a directory, false otherwise.
//     The path is resolved relative to the workingDirectory.
func setupJavascriptVM(workingDirectory string) *sobek.Runtime {
	vm := sobek.New()

	// Create the mxlint object
	mxlint := vm.NewObject()
	vm.Set("mxlint", mxlint)

	// Set the readfile function
	mxlint.Set("readfile", func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewGoError(fmt.Errorf("mxlint.readfile requires a file path argument")))
		}
		filepathArg := call.Argument(0).String()

		absPath, err := resolvePath(filepathArg, workingDirectory)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("mxlint.readfile: %w", err)))
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return vm.ToValue(string(content))
	})

	// Set the listdir function
	mxlint.Set("listdir", func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewGoError(fmt.Errorf("mxlint.listdir requires a directory path argument")))
		}
		dirpathArg := call.Argument(0).String()

		absPath, err := resolvePath(dirpathArg, workingDirectory)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("mxlint.listdir: %w", err)))
		}

		entries, err := os.ReadDir(absPath)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Convert directory entries to a slice of names
		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		return vm.ToValue(names)
	})

	// Set the isdir function
	mxlint.Set("isdir", func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewGoError(fmt.Errorf("mxlint.isdir requires a path argument")))
		}
		pathArg := call.Argument(0).String()

		absPath, err := resolvePath(pathArg, workingDirectory)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("mxlint.isdir: %w", err)))
		}

		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return vm.ToValue(false)
			}
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(info.IsDir())
	})

	return vm
}

func evalTestcase_Javascript(rulePath string, inputFilePath string, ruleNumber string, ignoreNoqa bool) (*Testcase, error) {
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

	// Use the directory containing the input file as the working directory
	workingDirectory := filepath.Dir(inputFilePath)
	vm := setupJavascriptVM(workingDirectory)
	_, err = vm.RunString(string(ruleContent))
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
