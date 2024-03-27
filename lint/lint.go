package lint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/open-policy-agent/opa/rego"
)

func EvalAll(policiesPath string, modelSourcePath string) (PolicyResult, error) {
	filepath.Walk(policiesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".rego") {
			Eval(path, modelSourcePath)
		}
		return nil
	})
	return PolicyResultPass, nil
}

func Eval(policyPath string, modelSourcePath string) (PolicyResult, error) {

	log.Debugf("evaluating policy %s", policyPath)

	// read the policy file
	policyFile, err := os.Open(policyPath)
	if err != nil {
		return PolicyResultUnknown, err
	}
	defer policyFile.Close()

	policyContent, err := os.ReadFile(policyPath)
	if err != nil {
		return PolicyResultUnknown, err
	}
	var inputFiles []string = nil
	var packageName string = ""
	var pattern string = ""
	var policy_canonical_name string = ""

	lines := strings.Split(string(policyContent), "\n")

	for _, line := range lines {
		tokens := strings.Split(line, "# input: ")
		if len(tokens) > 1 && inputFiles == nil {
			pattern = tokens[1]
			inputFiles, err = expandPaths(pattern, modelSourcePath)
			if err != nil {
				return PolicyResultUnknown, err
			}
		}
		tokens = strings.Split(line, "package ")
		if len(tokens) > 1 && packageName == "" {
			packageName = tokens[1]
		}
		tokens = strings.Split(line, "default ")
		if len(tokens) > 1 && policy_canonical_name == "" {
			policy_canonical_name = strings.Split(tokens[1], " := ")[0]
		}
	}

	log.Debugf("package name: %s", packageName)
	log.Debugf("policy name: %s", policy_canonical_name)
	log.Debugf("input pattern: %s", pattern)
	log.Debugf("expanded input files %v", inputFiles)

	queryString := "data." + packageName + "." + policy_canonical_name + " == true"

	results := make([]PolicyResult, 0)

	for _, inputFile := range inputFiles {
		result, err := evalSingle(policyPath, queryString, inputFile)
		resultString := "pass"
		if result == PolicyResultFail {
			resultString = "fail"
		} else if result == PolicyResultUnknown {
			resultString = "unknown"
		}
		fmt.Printf("%s \t%s.%s @ %s\n", resultString, packageName, policy_canonical_name, inputFile)
		if err != nil {
			return PolicyResultUnknown, err
		}
		results = append(results, result)
	}

	for _, result := range results {
		if result == PolicyResultFail {
			return PolicyResultFail, nil
		}
	}

	for _, result := range results {
		if result == PolicyResultUnknown {
			return PolicyResultUnknown, nil
		}
	}
	return PolicyResultPass, nil
}

func evalSingle(policyPath string, queryString string, inputFilePath string) (PolicyResult, error) {
	regoFile, _ := os.ReadFile(policyPath)
	log.Debugf("rego file: \n%s", regoFile)

	yamlFile, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Errorf("Error reading YAML file: %s\n", err)
		return PolicyResultUnknown, err
	}

	var data map[string]interface{}
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		log.Errorf("Error parsing YAML file: %s\n", err)
		return PolicyResultUnknown, err
	}

	ctx := context.Background()

	r := rego.New(
		rego.Query(queryString),
		rego.Load([]string{policyPath}, nil),
		rego.Input(data),
		rego.Trace(true),
	)

	rs, err := r.Eval(ctx)
	if err != nil {
		log.Fatal(err)
		return PolicyResultUnknown, err
	}

	log.Debugf("Result: %v", rs)
	if len(rs) == 0 {
		return PolicyResultUnknown, nil
	}
	result := rs[0].Expressions[0].Value.(bool)
	if !result {
		// fmt.Printf("policy failed for %s @ %s\n", policyPath, inputFilePath)
		// rego.PrintTraceWithLocation(os.Stdout, r)
		return PolicyResultFail, nil
	}
	return PolicyResultPass, nil
}
