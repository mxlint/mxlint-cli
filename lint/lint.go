package lint

import (
	"context"
	"encoding/xml"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/open-policy-agent/opa/rego"
)

func printTestsuite(ts Testsuite) {
	fmt.Printf("## %s\n", ts.Name)
	for _, tc := range ts.Testcases {
		result := "PASS"
		if tc.Failure != nil {
			result = "FAIL"
		}
		if tc.Skipped != nil {
			result = "SKIP"
		}
		fmt.Printf("%s (%.5fs) %s\n", result, tc.Time, tc.Name)
	}
	fmt.Println("")
}

func EvalAll(policiesPath string, modelSourcePath string, xunitReport string, jsonFile string) error {
	testsuites := make([]Testsuite, 0)
	policies, err := readPoliciesMetadata(policiesPath)
	if err != nil {
		return err
	}
	failuresCount := 0
	for _, policy := range policies {
			testsuite, err := evalTestsuite(policy, modelSourcePath)
			if err != nil {
				return err
			}
			printTestsuite(*testsuite)
			failuresCount += testsuite.Failures
			testsuites = append(testsuites, *testsuite)
	}

	if xunitReport != "" {
		file, err := os.Create(xunitReport)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		encoder := xml.NewEncoder(file)
		encoder.Indent("", "  ")
		testsuitesContainer := TestSuites{Testsuites: testsuites}
		if err := encoder.Encode(testsuitesContainer); err != nil {
			panic(err)
		}
	}

	if jsonFile != "" {
		file, err := os.Create(jsonFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		testsuitesContainer := TestSuites{Testsuites: testsuites, Policies: policies}
		if err := encoder.Encode(testsuitesContainer); err != nil {
			panic(err)
		}
	}

	if failuresCount > 0 {
		return fmt.Errorf("%d failures", failuresCount)
	}
	return nil
}

func evalTestsuite(policy Policy, modelSourcePath string) (*Testsuite, error) {

	log.Debugf("evaluating policy %s", policy.Path)

	var skipped *Skipped = nil
	if policy.SkipReason != "" {
		skipped = &Skipped{
			Message: policy.SkipReason,
		}
	}

	queryString := "data." + policy.PackageName
	testcases := make([]Testcase, 0)
	failuresCount := 0
	skippedCount := 0
	totalTime := 0.0
	inputFiles, err := expandPaths(policy.Pattern, modelSourcePath)
	if err != nil {
			return nil, err
	}
	testcase := &Testcase{}

	for _, inputFile := range inputFiles {
		if skipped != nil {
			testcase = &Testcase{
				Name:    inputFile,
				Time:    0,
				Skipped: skipped,
			}
			skippedCount++
		} else {
			testcase, err = evalTestcase(policy.Path, queryString, inputFile)
			if err != nil {
				return nil, err
			}
		}
		if testcase.Failure != nil {
			failuresCount++
		}
		totalTime += testcase.Time

		testcases = append(testcases, *testcase)
	}

	testsuite := &Testsuite{
		Name:      policy.Path,
		Tests:     len(testcases),
		Failures:  failuresCount,
		Skipped:   skippedCount,
		Time:      totalTime,
		Testcases: testcases,
	}

	return testsuite, nil
}

func evalTestcase(policyPath string, queryString string, inputFilePath string) (*Testcase, error) {
	regoFile, _ := os.ReadFile(policyPath)
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

	ctx := context.Background()

	startTime := time.Now()
	r := rego.New(
		rego.Query(queryString),
		rego.Load([]string{policyPath}, nil),
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
			log.Warnf("Rule failed: %s", err)
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
