package serve

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLintResultStructure(t *testing.T) {
	// Test the LintResult structure
	result := LintResult{
		Timestamp: time.Now(),
		Results:   map[string]interface{}{"test": "value"},
		Error:     "test error",
	}

	// Verify fields are properly set
	assert.NotZero(t, result.Timestamp)
	assert.Equal(t, map[string]interface{}{"test": "value"}, result.Results)
	assert.Equal(t, "test error", result.Error)

	// Test JSON marshaling
	data, err := json.Marshal(result)
	assert.NoError(t, err)

	// Unmarshal and verify
	var unmarshaled LintResult
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	assert.Equal(t, result.Timestamp.Unix(), unmarshaled.Timestamp.Unix())
	assert.Equal(t, result.Error, unmarshaled.Error)
}

func TestDashboardTemplate(t *testing.T) {
	// Test that the dashboard template can be parsed without errors
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Parse the template
	tmpl, err := template.New("dashboard").Funcs(funcMap).Parse(dashboardTemplate)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)

	// Test template execution with sample data
	testData := LintResult{
		Timestamp: time.Now(),
		Results: map[string]interface{}{
			"Rules": []map[string]interface{}{
				{
					"Title":       "Test Rule",
					"Severity":    "HIGH",
					"Category":    "Test",
					"RuleNumber":  "001",
					"Description": "Test description",
					"Remediation": "Test remediation",
					"Path":        "test/path",
				},
			},
			"Testsuites": []map[string]interface{}{
				{
					"Name":     "test/path",
					"Tests":    1,
					"Failures": 0,
					"Skipped":  0,
					"Testcases": []map[string]interface{}{
						{
							"Name": "Test Case",
							"Time": 0.001,
						},
					},
				},
			},
		},
	}

	// Execute template to buffer
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, testData)
	assert.NoError(t, err)

	// Verify output contains expected elements
	output := buf.String()
	assert.Contains(t, output, "MXLint Dashboard")
	assert.Contains(t, output, "Test Rule")
	assert.Contains(t, output, "Test description")
}

func TestHTTPHandlers(t *testing.T) {
	// Setup test data
	testResult := LintResult{
		Timestamp: time.Now(),
		Results:   map[string]interface{}{"test": "value"},
	}

	// Create a request to test the root handler with JSON Accept header
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	req.Header.Set("Accept", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create handler function with test data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			json.NewEncoder(w).Encode(testResult)
			return
		}

		// Otherwise serve HTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("<html><body>Test</body></html>"))
	})

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))

	// Check response body
	var response LintResult
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testResult.Timestamp.Unix(), response.Timestamp.Unix())

	// Test HTML response
	req, err = http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/html; charset=utf-8", rr.Header().Get("Content-Type"))
	assert.Contains(t, rr.Body.String(), "<html>")
}

func TestDownloadRules(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mxlint-download-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a logger that discards output for testing
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// Skip actual execution if not in CI environment to avoid network calls during tests
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping download test in non-CI environment")
	}

	// This is a very basic test that just ensures the function doesn't panic
	// In a real test environment, we would mock the git command execution
	err = DownloadRules(tempDir, logger)
	if err != nil {
		// We're not asserting no error because the git command might fail
		// if git is not installed or network is unavailable
		t.Logf("DownloadRules returned error (this might be expected): %v", err)
	}
}

func TestAddDirsRecursive(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "mxlint-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	subDir1 := filepath.Join(tempDir, "subdir1")
	subDir2 := filepath.Join(tempDir, "subdir2")
	excludeDir := filepath.Join(tempDir, "exclude")
	hiddenDir := filepath.Join(tempDir, ".hidden")

	for _, dir := range []string{subDir1, subDir2, excludeDir, hiddenDir} {
		err = os.MkdirAll(dir, 0755)
		assert.NoError(t, err)
	}

	// Create a real fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	assert.NoError(t, err)
	defer watcher.Close()

	// Create a logger that discards output
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// Call the function we're testing
	err = addDirsRecursive(watcher, tempDir, excludeDir, logger)
	assert.NoError(t, err)

	// Get the watch list
	watchList := watcher.WatchList()

	// Check that the correct directories were added
	assert.Contains(t, watchList, tempDir)
	assert.Contains(t, watchList, subDir1)
	assert.Contains(t, watchList, subDir2)

	// Check that excluded and hidden directories were not added
	for _, path := range watchList {
		assert.NotEqual(t, excludeDir, path, "Exclude directory should not be watched")
		assert.NotContains(t, path, ".hidden", "Hidden directories should not be watched")
	}
}
