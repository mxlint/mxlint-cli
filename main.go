package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mxlint/mxlint-cli/lint"
	"github.com/mxlint/mxlint-cli/mpr"
	"github.com/mxlint/mxlint-cli/serve"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var bakedDefaultConfigYAML []byte

func main() {
	lint.SetDefaultConfigYAML(bakedDefaultConfigYAML)

	var rootCmd = &cobra.Command{Use: "mxlint-cli"}
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Turn on debug logs for all commands")
	rootCmd.PersistentFlags().String("config", "", "Path to config file (highest precedence)")

	var cmdExportModel = &cobra.Command{
		Use:   "export-model",
		Short: "Export Mendix model to yaml files",
		Long:  "The output is a text representation of the model. It is a one-way conversion that aims to keep the semantics yet readable for humans and computers.",
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("failed to resolve current working directory: %s\n", err)
				os.Exit(1)
			}

			config, err := lint.LoadMergedConfigFromPath(projectDir, configPathForCommand(cmd))
			if err != nil {
				fmt.Printf("failed to load configuration: %s\n", err)
				os.Exit(1)
			}
			log := logrus.New()
			if isVerbose(cmd) {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}
			mpr.SetLogger(log)

			inputDirectory := config.ProjectDirectory
			outputDirectory := config.Modelsource

			err = mpr.ExportModel(
				inputDirectory,
				outputDirectory,
				boolValue(config.Export.Raw, false),
				config.Export.Mode,
				boolValue(config.Export.Appstore, false),
				config.Export.Filter,
			)
			if err != nil {
				log.Errorf("export-model failed: %s", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmdExportModel)

	var cmdLint = &cobra.Command{
		Use:   "lint",
		Short: "Evaluate Mendix model against rules. Requires the model to be exported first",
		Long:  "The model is evaluated against a set of rules. The rules are defined in OPA rego files. The output is a list of checked rules and their outcome.",
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("failed to resolve current working directory: %s\n", err)
				os.Exit(1)
			}

			config, err := lint.LoadMergedConfigFromPath(projectDir, configPathForCommand(cmd))
			if err != nil {
				fmt.Printf("failed to load configuration: %s\n", err)
				os.Exit(1)
			}
			log := logrus.New()
			if isVerbose(cmd) {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}
			lint.SetLogger(log)
			lint.SetConfig(config)

			rulesDirectory := config.Rules.Path
			modelDirectory := config.Modelsource

			if !filepath.IsAbs(rulesDirectory) {
				rulesDirectory = filepath.Join(projectDir, rulesDirectory)
			}

			if config != nil && len(config.Rules.Rulesets) > 0 {
				log.Infof("Syncing %d rulesets to %s", len(config.Rules.Rulesets), rulesDirectory)
				if err := lint.SyncRulesets(config.Rules.Rulesets, rulesDirectory, projectDir); err != nil {
					log.Errorf("failed to sync rulesets: %s", err)
					os.Exit(1)
				}
			}

			err = lint.EvalAll(
				rulesDirectory,
				modelDirectory,
				config.Lint.XunitReport,
				config.Lint.JSONFile,
				boolValue(config.Lint.IgnoreNoqa, false),
				!boolValue(config.Lint.NoCache, false),
			)
			if err != nil {
				log.Errorf("lint failed: %s", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmdLint)

	var cmdConfig = &cobra.Command{
		Use:   "config",
		Short: "Show merged active configuration",
		Long:  "Shows the merged active configuration and which config sources were found and used.",
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("failed to resolve current working directory: %s\n", err)
				os.Exit(1)
			}

			config, report, err := lint.LoadMergedConfigWithReportFromPath(projectDir, configPathForCommand(cmd))
			if err != nil {
				fmt.Printf("failed to load configuration: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("Config Sources")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SOURCE\tFOUND\tUSED\tPATH")
			fmt.Fprintf(w, "%s\t%t\t%t\t%s\n", report.Default.Name, report.Default.Found, report.Default.Used, report.Default.Path)
			fmt.Fprintf(w, "%s\t%t\t%t\t%s\n", report.System.Name, report.System.Found, report.System.Used, report.System.Path)
			fmt.Fprintf(w, "%s\t%t\t%t\t%s\n", report.Project.Name, report.Project.Found, report.Project.Used, report.Project.Path)
			fmt.Fprintf(w, "%s\t%t\t%t\t%s\n", report.Explicit.Name, report.Explicit.Found, report.Explicit.Used, report.Explicit.Path)
			_ = w.Flush()

			yamlBytes, err := yaml.Marshal(config)
			if err != nil {
				fmt.Printf("failed to marshal merged configuration: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("\nMerged Active Configuration")
			fmt.Print(string(yamlBytes))
		},
	}
	rootCmd.AddCommand(cmdConfig)

	// Add the serve command
	serveCmd := serve.NewServeCommand()
	rootCmd.AddCommand(serveCmd)

	var cmdRules = &cobra.Command{
		Use:   "test-rules",
		Short: "Ensure rules are working as expected against predefined test cases",
		Long:  "When you are developing a new rule, you can use this command to ensure it works as expected against predefined test cases.",
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := os.Getwd()
			if err != nil {
				fmt.Printf("failed to resolve current working directory: %s\n", err)
				os.Exit(1)
			}
			config, err := lint.LoadMergedConfigFromPath(projectDir, configPathForCommand(cmd))
			if err != nil {
				fmt.Printf("failed to load configuration: %s\n", err)
				os.Exit(1)
			}
			log := logrus.New()
			if isVerbose(cmd) {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}
			lint.SetLogger(log)
			err = lint.TestAll(config.Rules.Path)
			if err != nil {
				log.Errorf("Test rules failed: %s", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmdRules)

	var cmdCacheClear = &cobra.Command{
		Use:   "cache-clear",
		Short: "Clear the lint results cache",
		Long:  "Removes all cached lint results. The cache is used to speed up repeated linting operations when rules and model files haven't changed.",
		Run: func(cmd *cobra.Command, args []string) {
			log := logrus.New()
			if isVerbose(cmd) {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}
			lint.SetLogger(log)
			err := lint.ClearCache()
			if err != nil {
				log.Errorf("Failed to clear cache: %s", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(cmdCacheClear)

	var cmdCacheStats = &cobra.Command{
		Use:   "cache-stats",
		Short: "Show cache statistics",
		Long:  "Displays information about the cached lint results, including number of entries and total size.",
		Run: func(cmd *cobra.Command, args []string) {
			log := logrus.New()
			if isVerbose(cmd) {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}
			lint.SetLogger(log)
			count, size, err := lint.GetCacheStats()
			if err != nil {
				log.Errorf("Failed to get cache stats: %s", err)
				os.Exit(1)
			}

			sizeInKB := float64(size) / 1024.0
			sizeInMB := sizeInKB / 1024.0

			log.Infof("Cache Statistics:")
			log.Infof("  Entries: %d", count)
			if sizeInMB >= 1.0 {
				log.Infof("  Total Size: %.2f MB", sizeInMB)
			} else {
				log.Infof("  Total Size: %.2f KB", sizeInKB)
			}
		},
	}
	rootCmd.AddCommand(cmdCacheStats)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func isVerbose(cmd *cobra.Command) bool {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return false
	}
	return verbose
}

func configPathForCommand(cmd *cobra.Command) string {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return ""
	}
	return configPath
}
