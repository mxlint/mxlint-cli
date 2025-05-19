package lint

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	rules, err := ReadRulesMetadata(rulesPath)
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

		if rule.Language == LanguageRego {
			testcase, err = evalTestcase_Rego(rule.Path, queryString, inputFile)
		} else if rule.Language == LanguageJavascript {
			testcase, err = evalTestcase_Javascript(rule.Path, inputFile)
		}
		if err != nil {
			return nil, err
		}
		if testcase.Failure != nil {
			failuresCount++
		}

		if testcase.Skipped != nil {
			skippedCount++
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

func ReadRulesMetadata(rulesPath string) ([]Rule, error) {
	rules := make([]Rule, 0)
	filepath.Walk(rulesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(info.Name(), "_test.rego") && strings.HasSuffix(info.Name(), ".rego") {
			rule, err := parseRuleMetadata_Rego(path)
			if err != nil {
				return err
			}
			rules = append(rules, *rule)
		}
		if !info.IsDir() && !strings.HasSuffix(info.Name(), "_test.js") && strings.HasSuffix(info.Name(), ".js") {
			rule, err := parseRuleMetadata_Javascript(path)
			if err != nil {
				return err
			}
			rules = append(rules, *rule)
		}
		return nil
	})
	return rules, nil
}
