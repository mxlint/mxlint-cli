package lint

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
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

func EvalAll(policiesPath string, modelSourcePath string, xunitReport string) error {
	testsuites := make([]Testsuite, 0)
	failuresCount := 0
	filepath.Walk(policiesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(info.Name(), "_test.rego") && strings.HasSuffix(info.Name(), ".rego") {
			testsuite, err := evalTestsuite(path, modelSourcePath)
			if err != nil {
				return err
			}
			printTestsuite(*testsuite)
			failuresCount += testsuite.Failures
			testsuites = append(testsuites, *testsuite)
		}
		return nil
	})

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

	if failuresCount > 0 {
		return fmt.Errorf("%d failures", failuresCount)
	}
	return nil
}

func evalTestsuite(policyPath string, modelSourcePath string) (*Testsuite, error) {

	log.Debugf("evaluating policy %s", policyPath)

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
	var inputFiles []string = nil
	var packageName string = ""
	var pattern string = ""
	var policy_canonical_name string = ""
	var skipReason string = ""

	lines := strings.Split(string(policyContent), "\n")

	for _, line := range lines {
		tokens := strings.Split(line, "#  input: ")
		if len(tokens) > 1 && inputFiles == nil {
			pattern = strings.ReplaceAll(tokens[1], "\"", "")
			inputFiles, err = expandPaths(pattern, modelSourcePath)
			if err != nil {
				return nil, err
			}
		}
		tokens = strings.Split(line, "#  skip: ")
		if len(tokens) > 1 && skipReason == "" {
			skipReason = tokens[1]
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

	var skipped *Skipped = nil
	if skipReason != "" {
		skipped = &Skipped{
			Message: skipReason,
		}
	}

	queryString := "data." + packageName
	testcases := make([]Testcase, 0)
	failuresCount := 0
	skippedCount := 0
	totalTime := 0.0
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
			testcase, err = evalTestcase(policyPath, queryString, inputFile)
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
		Name:      policyPath,
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
		buf := new(strings.Builder)
		rego.PrintTraceWithLocation(buf, r)
		myErrors := make([]string, 0)
		for _, err := range errors {
			log.Warnf("Rule failed: %s", err)
			myErrors = append(myErrors, fmt.Sprintf("%s", err))
		}
		failure = &Failure{
			Message: strings.Join(myErrors, "\n"),
			Type:    "AssertionError",
			Data:    buf.String(),
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
