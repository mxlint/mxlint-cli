package serve

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mxlint/mxlint-cli/lint"
	"github.com/mxlint/mxlint-cli/mpr"
	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// LintResult stores the results of linting
type LintResult struct {
	Timestamp time.Time   `json:"timestamp"`
	Results   interface{} `json:"results"`
	Error     string      `json:"error,omitempty"`
}

// NewServeCommand creates a new serve command
func NewServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run a server that exports model and lints whenever the input MPR file changes",
		Long:  "This command runs export-model followed by linting whenever the input MPR file changes. It's useful for continuous linting during development.",
		Run:   runServe,
	}

	cmd.Flags().StringP("input", "i", ".", "Path to directory or mpr file to export. If it's a directory, all mpr files will be exported")
	cmd.Flags().StringP("output", "o", "modelsource", "Path to directory to write the yaml files. If it doesn't exist, it will be created")
	cmd.Flags().StringP("mode", "m", "basic", "Export mode. Valid options: basic, advanced")
	cmd.Flags().StringP("rules", "r", "rules", "Path to directory with rules")
	cmd.Flags().IntP("port", "p", 8082, "Port to run the server on")
	cmd.Flags().Bool("verbose", false, "Turn on for debug logs")
	cmd.Flags().IntP("debounce", "d", 500, "Debounce time in milliseconds for file change events")

	return cmd
}

// runServe implements the serve command functionality
func runServe(cmd *cobra.Command, args []string) {
	inputDirectory, _ := cmd.Flags().GetString("input")
	outputDirectory, _ := cmd.Flags().GetString("output")
	mode, _ := cmd.Flags().GetString("mode")
	rulesDirectory, _ := cmd.Flags().GetString("rules")
	port, _ := cmd.Flags().GetInt("port")
	verbose, _ := cmd.Flags().GetBool("verbose")
	debounceTime, _ := cmd.Flags().GetInt("debounce")

	w := watcher.New()
	w.IgnoreHiddenFiles(true)

	log := logrus.New()
	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	mpr.SetLogger(log)
	lint.SetLogger(log)

	// Check if rules directory exists, if not download it
	if _, err := os.Stat(rulesDirectory); os.IsNotExist(err) {
		if err := DownloadRules(rulesDirectory, log); err != nil {
			log.Fatalf("Failed to download rules: %v", err)
		}
	}

	expandedPath, err := filepath.Abs(inputDirectory)
	if err != nil {
		log.Fatalln(err)
	}

	log.Infof("Starting server on port %d", port)
	log.Infof("Watching for changes in %s", expandedPath)
	log.Infof("Output directory: %s", outputDirectory)
	log.Infof("Rules directory: %s", rulesDirectory)
	log.Infof("Mode: %s", mode)
	log.Infof("Debounce time: %d ms", debounceTime)

	// Create a mutex to protect the cached results
	var resultMutex sync.RWMutex
	var cachedResult LintResult

	// Create template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	// Parse the HTML template
	tmpl, err := template.New("dashboard").Funcs(funcMap).Parse(dashboardTemplate)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	// Set up HTTP server to serve the lint results
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resultMutex.RLock()
		defer resultMutex.RUnlock()

		// Check if the client wants JSON
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")

			if cachedResult.Timestamp.IsZero() {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{
					"status": "No lint results available yet. Please try again later.",
				})
				return
			}

			json.NewEncoder(w).Encode(cachedResult)
			return
		}

		// Otherwise serve HTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if cachedResult.Timestamp.IsZero() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("<html><body><h1>No lint results available yet</h1><p>Please try again later.</p></body></html>"))
			return
		}

		if err := tmpl.Execute(w, cachedResult); err != nil {
			log.Errorf("Error executing template: %v", err)
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
	})

	// Add API endpoint for JSON results
	http.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		resultMutex.RLock()
		defer resultMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if cachedResult.Timestamp.IsZero() {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "No lint results available yet. Please try again later.",
			})
			return
		}

		json.NewEncoder(w).Encode(cachedResult)
	})

	// Start HTTP server in a goroutine
	go func() {
		serverAddr := fmt.Sprintf("127.0.0.1:%d", port)
		log.Infof("HTTP server listening on %s", serverAddr)
		log.Infof("Dashboard available at http://localhost:%d", port)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Function to run export and lint
	runExportAndLint := func() {
		// Use defer to recover from any panics
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic in runExportAndLint: %v", r)
				resultMutex.Lock()
				cachedResult = LintResult{
					Timestamp: time.Now(),
					Error:     fmt.Sprintf("Internal error: %v", r),
				}
				resultMutex.Unlock()
			}
		}()

		log.Infof("Running export-model and lint")
		err := mpr.ExportModel(inputDirectory, outputDirectory, false, mode, false)
		if err != nil {
			log.Warningf("Export failed: %s", err)
			resultMutex.Lock()
			cachedResult = LintResult{
				Timestamp: time.Now(),
				Error:     fmt.Sprintf("Export failed: %s", err),
			}
			resultMutex.Unlock()
			return
		}

		// Run lint and update cached results
		resultMutex.Lock()
		defer resultMutex.Unlock() // Ensure mutex is always unlocked

		var results interface{}
		var lintErr error

		// Wrap the lint operation in its own recover block
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("Recovered from panic in lint operation: %v", r)
					lintErr = fmt.Errorf("lint operation panicked: %v", r)
				}
			}()
			results, lintErr = lint.EvalAllWithResults(rulesDirectory, outputDirectory, "", "")
		}()

		if lintErr != nil {
			log.Warningf("Lint failed: %s", lintErr)
		}
		cachedResult = LintResult{
			Timestamp: time.Now(),
			Results:   results,
		}

	}

	// Watch for changes and update cached results with debouncing
	go func() {
		var timer *time.Timer
		var timerMutex sync.Mutex

		for {
			select {
			case event := <-w.Event:
				if verbose {
					log.Debugf("Change detected: %s", event)
				}

				timerMutex.Lock()
				// Cancel existing timer if it's running
				if timer != nil {
					timer.Stop()
				}

				// Create a new timer
				timer = time.AfterFunc(time.Duration(debounceTime)*time.Millisecond, func() {
					// Recover from any panics in the timer function
					defer func() {
						if r := recover(); r != nil {
							log.Errorf("Recovered from panic in timer function: %v", r)
						}
					}()
					runExportAndLint()
				})
				timerMutex.Unlock()

			case err := <-w.Error:
				log.Errorf("Watcher error: %v", err)
				// Don't fatal here, just log the error
			case <-w.Closed:
				return
			}
		}
	}()

	// Add directories to watch with error handling
	if err := w.AddRecursive(inputDirectory); err != nil {
		log.Errorf("Error adding directory to watch: %v", err)
		// Continue execution, don't fatal
	}

	// Ignore output directory
	w.Ignore(outputDirectory)

	// first run
	go func() {
		// Recover from any panics in the initial run
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic in initial run: %v", r)
				resultMutex.Lock()
				cachedResult = LintResult{
					Timestamp: time.Now(),
					Error:     fmt.Sprintf("Initial run failed with panic: %v", r),
				}
				resultMutex.Unlock()
			}
		}()

		w.Wait()
		log.Info("Initial export and lint")
		runExportAndLint()
	}()

	// Start the watcher with error handling
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Errorf("Error starting watcher: %v", err)
		// Instead of fatal, set an error in the cached result
		resultMutex.Lock()
		cachedResult = LintResult{
			Timestamp: time.Now(),
			Error:     fmt.Sprintf("Failed to start file watcher: %v", err),
		}
		resultMutex.Unlock()
	}
}
