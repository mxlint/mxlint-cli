package lint

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/open-policy-agent/opa/rego"
)

const NOQA = "# noqa"
const NOQA_ALIAS = "#noqa"

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

func EvalAll(rulesPath string, modelSourcePath string, xunitReport string, jsonFile string) error {
	testsuites := make([]Testsuite, 0)
	rules, err := readRulesMetadata(rulesPath)
	if err != nil {
		return err
	}
	failuresCount := 0
	for _, rule := range rules {
		testsuite, err := evalTestsuite(rule, modelSourcePath)
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
		testsuitesContainer := TestSuites{Testsuites: testsuites, Rules: rules}
		if err := encoder.Encode(testsuitesContainer); err != nil {
			panic(err)
		}
	}

	for _, ts := range testsuites {
		if ts.Failures > 0 {
			log.Warningf("Rule %s: %d failures", ts.Name, ts.Failures)
			for _, tc := range ts.Testcases {
				if tc.Failure != nil {
					log.Warningf("  Document %s: %s", tc.Name, tc.Failure.Message)
				}
			}
		}
	}

	if failuresCount > 0 {
		return fmt.Errorf("%d failures", failuresCount)
	} else {
		log.Infof("All good my friend")
	}
	return nil
}

func evalTestsuite(rule Rule, modelSourcePath string) (*Testsuite, error) {

	log.Debugf("evaluating rule %s", rule.Path)

	var skipped *Skipped = nil
	if rule.SkipReason != "" {
		skipped = &Skipped{
			Message: rule.SkipReason,
		}
	}

	queryString := "data." + rule.PackageName
	testcases := make([]Testcase, 0)
	failuresCount := 0
	skippedCount := 0
	totalTime := 0.0
	inputFiles, err := expandPaths(rule.Pattern, modelSourcePath)
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
			testcase, err = evalTestcase(rule.Path, queryString, inputFile)
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
		Name:      rule.Path,
		Tests:     len(testcases),
		Failures:  failuresCount,
		Skipped:   skippedCount,
		Time:      totalTime,
		Testcases: testcases,
	}

	return testsuite, nil
}

func evalTestcase(rulePath string, queryString string, inputFilePath string) (*Testcase, error) {
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
			if strings.HasPrefix(line, NOQA) || strings.HasPrefix(line, NOQA_ALIAS) {
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
