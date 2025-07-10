package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mxlint/mxlint-cli/lint"
	"github.com/mxlint/mxlint-cli/mpr"
	"github.com/mxlint/mxlint-cli/serve"
	"github.com/radovskyb/watcher"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {

	var rootCmd = &cobra.Command{Use: "mxlint-cli"}

	var cmdExportModel = &cobra.Command{
		Use:   "export-model",
		Short: "Export Mendix model to yaml files",
		Long:  "The output is a text representation of the model. It is a one-way conversion that aims to keep the semantics yet readable for humans and computers.",
		Run: func(cmd *cobra.Command, args []string) {
			inputDirectory, _ := cmd.Flags().GetString("input")
			outputDirectory, _ := cmd.Flags().GetString("output")
			raw, _ := cmd.Flags().GetBool("raw")
			mode, _ := cmd.Flags().GetString("mode")
			verbose, _ := cmd.Flags().GetBool("verbose")
			appstore, _ := cmd.Flags().GetBool("appstore")

			log := logrus.New()
			if verbose {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}

			mpr.SetLogger(log)
			mpr.ExportModel(inputDirectory, outputDirectory, raw, mode, appstore)
		},
	}

	cmdExportModel.Flags().StringP("input", "i", ".", "Path to directory or mpr file to export. If it's a directory, all mpr files will be exported")
	cmdExportModel.Flags().StringP("output", "o", "modelsource", "Path to directory to write the yaml files. If it doesn't exist, it will be created")
	cmdExportModel.Flags().StringP("mode", "m", "basic", "Export mode. Valid options: basic, advanced")
	cmdExportModel.Flags().Bool("raw", false, "If set, the output yaml will include all attributes as they are in the model. Otherwise, only the relevant attributes are included. You should never need this. Only useful when you are developing new functionalities for this tool.")
	cmdExportModel.Flags().Bool("appstore", false, "If set, appstore modules will be included in the output")
	cmdExportModel.Flags().Bool("verbose", false, "Turn on for debug logs")
	rootCmd.AddCommand(cmdExportModel)

	var cmdLint = &cobra.Command{
		Use:   "lint",
		Short: "Evaluate Mendix model against rules. Requires the model to be exported first",
		Long:  "The model is evaluated against a set of rules. The rules are defined in OPA rego files. The output is a list of checked rules and their outcome.",
		Run: func(cmd *cobra.Command, args []string) {
			rulesDirectory, _ := cmd.Flags().GetString("rules")
			modelDirectory, _ := cmd.Flags().GetString("modelsource")
			xunitReport, _ := cmd.Flags().GetString("xunit-report")
			JsonFile, _ := cmd.Flags().GetString("json-file")
			verbose, _ := cmd.Flags().GetBool("verbose")

			log := logrus.New()
			if verbose {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}

			lint.SetLogger(log)
			err := lint.EvalAll(rulesDirectory, modelDirectory, xunitReport, JsonFile)
			if err != nil {
				log.Errorf("lint failed: %s", err)
				os.Exit(1)
			}
		},
	}

	cmdLint.Flags().StringP("rules", "r", "rules", "Path to directory with rules")
	cmdLint.Flags().StringP("modelsource", "m", "modelsource", "Path to directory with exported model")
	cmdLint.Flags().StringP("xunit-report", "x", "", "Path to output file for xunit report. If not provided, no xunit report will be generated")
	cmdLint.Flags().StringP("json-file", "j", "", "Path to output file for JSON report. If not provided, no JSON file will be generated")
	cmdLint.Flags().Bool("verbose", false, "Turn on for debug logs")
	rootCmd.AddCommand(cmdLint)

	var cmdWatch = &cobra.Command{
		Use:   "watch",
		Short: "Watch for changes in the model, export-model and lint continuously",
		Long:  "Continuous linting of the model. This is useful when you are developing your application and want to be notified of any changes that might break the rules.",
		Run: func(cmd *cobra.Command, args []string) {
			inputDirectory, _ := cmd.Flags().GetString("input")
			outputDirectory, _ := cmd.Flags().GetString("output")
			mode, _ := cmd.Flags().GetString("mode")
			rulesDirectory, _ := cmd.Flags().GetString("rules")

			w := watcher.New()
			w.IgnoreHiddenFiles(true)

			log := logrus.New()
			log.SetLevel(logrus.InfoLevel)

			mpr.SetLogger(log)
			lint.SetLogger(log)

			expandedPath, err := filepath.Abs(inputDirectory)
			if err != nil {
				log.Fatalln(err)
			}

			go func() {
				for {
					select {
					case event := <-w.Event:
						fmt.Println(event)

						log.Infof("Watching for changes in %s", expandedPath)
						log.Infof("Output directory: %s", outputDirectory)
						log.Infof("Rules directory: %s", rulesDirectory)
						log.Infof("Mode: %s", mode)
						mpr.ExportModel(inputDirectory, outputDirectory, false, mode, false)
						err := lint.EvalAll(rulesDirectory, outputDirectory, "", "")
						if err != nil {
							log.Warningf("Lint failed: %s", err)
						}
					case err := <-w.Error:
						log.Fatalln(err)
					case <-w.Closed:
						return
					}
				}
			}()

			if err := w.AddRecursive(inputDirectory); err != nil {
				log.Fatalln(err)
			}
			w.Ignore(outputDirectory)

			// first run
			go func() {
				w.Wait()
				w.TriggerEvent(watcher.Create, nil)
			}()

			if err := w.Start(time.Millisecond * 100); err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmdWatch.Flags().StringP("input", "i", ".", "Path to directory or mpr file to export. If it's a directory, all mpr files will be exported")
	cmdWatch.Flags().StringP("output", "o", "modelsource", "Path to directory to write the yaml files. If it doesn't exist, it will be created")
	cmdWatch.Flags().StringP("mode", "m", "basic", "Export mode. Valid options: basic, advanced")
	cmdWatch.Flags().StringP("rules", "r", "rules", "Path to directory with rules")
	rootCmd.AddCommand(cmdWatch)

	// Add the serve command
	serveCmd := serve.NewServeCommand()
	rootCmd.AddCommand(serveCmd)

	var cmdRules = &cobra.Command{
		Use:   "test-rules",
		Short: "Ensure rules are working as expected against predefined test cases",
		Long:  "When you are developing a new rule, you can use this command to ensure it works as expected against predefined test cases.",
		Run: func(cmd *cobra.Command, args []string) {
			rulesDirectory, _ := cmd.Flags().GetString("rules")
			verbose, _ := cmd.Flags().GetBool("verbose")

			log := logrus.New()
			if verbose {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}

			lint.SetLogger(log)
			err := lint.TestAll(rulesDirectory)
			if err != nil {
				log.Errorf("Test rules failed: %s", err)
				os.Exit(1)
			}
		},
	}

	cmdRules.Flags().StringP("rules", "r", "rules", "Path to directory with rules")
	cmdRules.Flags().Bool("verbose", false, "Turn on for debug logs")
	rootCmd.AddCommand(cmdRules)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
