package lint

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// EvalAllWithResults evaluates all rules and returns the results
// This is similar to EvalAll but returns the results instead of just printing them
func EvalAllWithResults(rulesPath string, modelSourcePath string, xunitReport string, jsonFile string, ignoreNoqa bool) (interface{}, error) {
	rules, err := ReadRulesMetadata(rulesPath)
	if err != nil {
		return nil, err
	}

	// Create a slice to store results in order
	testsuites := make([]Testsuite, len(rules))

	// Use a WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Create a channel to collect errors
	errChan := make(chan error, len(rules))

	// Create a mutex to safely print testsuites
	var printMutex sync.Mutex

	// Launch goroutines to evaluate rules in parallel
	for i, rule := range rules {
		wg.Add(1)
		go func(index int, r Rule) {
			defer wg.Done()

			testsuite, err := evalTestsuite(r, modelSourcePath, ignoreNoqa)
			if err != nil {
				errChan <- err
				return
			}

			// Print with mutex to avoid interleaved output
			printMutex.Lock()
			printTestsuite(*testsuite)
			printMutex.Unlock()

			testsuites[index] = *testsuite
		}(i, rule)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	// Calculate total failures
	failuresCount := 0
	for _, ts := range testsuites {
		failuresCount += ts.Failures
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

	// Return the results
	testsuitesContainer := TestSuites{Testsuites: testsuites, Rules: rules}

	if failuresCount > 0 {
		return testsuitesContainer, fmt.Errorf("%d failures", failuresCount)
	} else {
		log.Infof("Lint summary: All rules passed successfully!")
		log.Infof("Total rules evaluated: %d", len(rules))
		log.Infof("Total files checked: %d", countTotalTestcases(testsuites))
	}
	return testsuitesContainer, nil
}

func EvalAll(rulesPath string, modelSourcePath string, xunitReport string, jsonFile string, ignoreNoqa bool) error {
	rules, err := ReadRulesMetadata(rulesPath)
	if err != nil {
		return err
	}

	// Create a slice to store results in order
	testsuites := make([]Testsuite, len(rules))

	// Use a WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Create a channel to collect errors
	errChan := make(chan error, len(rules))

	// Create a mutex to safely print testsuites
	var printMutex sync.Mutex

	// Launch goroutines to evaluate rules in parallel
	for i, rule := range rules {
		wg.Add(1)
		go func(index int, r Rule) {
			defer wg.Done()

			testsuite, err := evalTestsuite(r, modelSourcePath, ignoreNoqa)
			if err != nil {
				errChan <- err
				return
			}

			// Print with mutex to avoid interleaved output
			printMutex.Lock()
			printTestsuite(*testsuite)
			printMutex.Unlock()

			testsuites[index] = *testsuite
		}(i, rule)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	if len(errChan) > 0 {
		return <-errChan
	}

	// Calculate total failures
	failuresCount := 0
	for _, ts := range testsuites {
		failuresCount += ts.Failures
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
		log.Errorf("Lint summary: Found %d failures:", failuresCount)
		log.Errorf("Failures by rule:")
		for _, ts := range testsuites {
			if ts.Failures > 0 {
				log.Errorf("- %s: %d failures", ts.Name, ts.Failures)
			}
		}
		return fmt.Errorf("%d failures", failuresCount)
	} else {
		log.Infof("Lint summary: All rules passed successfully!")
		log.Infof("Total rules evaluated: %d", len(rules))
		log.Infof("Total files checked: %d", countTotalTestcases(testsuites))
	}
	return nil
}

// countTotalTestcases returns the total number of testcases across all testsuites
func countTotalTestcases(testsuites []Testsuite) int {
	count := 0
	for _, ts := range testsuites {
		count += len(ts.Testcases)
	}
	return count
}

func evalTestsuite(rule Rule, modelSourcePath string, ignoreNoqa bool) (*Testsuite, error) {

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

		// Try to load from cache first (but skip cache if ignoreNoqa is true)
		cacheKey, err := createCacheKey(rule.Path, inputFile)
		if err != nil {
			log.Debugf("Error creating cache key: %v", err)
		} else if !ignoreNoqa {
			cachedTestcase, found := loadCachedTestcase(*cacheKey)
			if found {
				testcase = cachedTestcase
				log.Debugf("Using cached result for %s", inputFile)
			} else {
				// Cache miss - evaluate and save to cache
				testcase, err = evalTestcaseWithCaching(rule, queryString, inputFile, cacheKey, ignoreNoqa)
				if err != nil {
					return nil, err
				}
			}
		} else {
			// ignoreNoqa is true, skip cache and evaluate directly
			testcase, err = evalTestcaseWithCaching(rule, queryString, inputFile, cacheKey, ignoreNoqa)
			if err != nil {
				return nil, err
			}
		}

		// Fallback if cache key creation failed
		if cacheKey == nil {
			if rule.Language == LanguageRego {
				testcase, err = evalTestcase_Rego(rule.Path, queryString, inputFile, rule.RuleNumber, ignoreNoqa)
			} else if rule.Language == LanguageJavascript {
				testcase, err = evalTestcase_Javascript(rule.Path, inputFile, rule.RuleNumber, ignoreNoqa)
			}
			if err != nil {
				return nil, err
			}
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

// evalTestcaseWithCaching evaluates a testcase and saves the result to cache
func evalTestcaseWithCaching(rule Rule, queryString string, inputFile string, cacheKey *CacheKey, ignoreNoqa bool) (*Testcase, error) {
	var testcase *Testcase
	var err error

	if rule.Language == LanguageRego {
		testcase, err = evalTestcase_Rego(rule.Path, queryString, inputFile, rule.RuleNumber, ignoreNoqa)
	} else if rule.Language == LanguageJavascript {
		testcase, err = evalTestcase_Javascript(rule.Path, inputFile, rule.RuleNumber, ignoreNoqa)
	}

	if err != nil {
		return nil, err
	}

	// Only save to cache when ignoreNoqa is false
	// When ignoreNoqa is true, the result might differ from the normal behavior
	if !ignoreNoqa {
		if cacheErr := saveCachedTestcase(*cacheKey, testcase); cacheErr != nil {
			log.Debugf("Error saving to cache: %v", cacheErr)
			// Don't fail the evaluation if cache save fails
		}
	}

	return testcase, nil
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
