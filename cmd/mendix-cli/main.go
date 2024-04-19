package main

import (
	"fmt"
	"os"

	"github.com/cinaq/mendix-cli/lint"
	"github.com/cinaq/mendix-cli/mpr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {

	var rootCmd = &cobra.Command{Use: "mendix-cli"}

	var cmdExportModel = &cobra.Command{
		Use:   "export-model ",
		Short: "Export Mendix model to yaml files",
		Long:  "The output is a text representation of the model. It is a one-way conversion that aims to keep the semantics yet readable for humans and computers.",
		Run: func(cmd *cobra.Command, args []string) {
			inputDirectory, _ := cmd.Flags().GetString("input")
			outputDirectory, _ := cmd.Flags().GetString("output")
			raw, _ := cmd.Flags().GetBool("raw")
			verbose, _ := cmd.Flags().GetBool("verbose")

			log := logrus.New()
			if verbose {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}

			mpr.SetLogger(log)
			mpr.ExportModel(inputDirectory, outputDirectory, raw)
		},
	}

	cmdExportModel.Flags().StringP("input", "i", ".", "Path to directory or mpr file to export. If it's a directory, all mpr files will be exported")
	cmdExportModel.Flags().StringP("output", "o", "modelsource", "Path to directory to write the yaml files. If it doesn't exist, it will be created")
	cmdExportModel.Flags().Bool("raw", false, "If set, the output yaml will include all attributes as they are in the model. Otherwise, only the relevant attributes are included. You should never need this. Only useful when you are developing new functionalities for this tool.")
	cmdExportModel.Flags().Bool("verbose", false, "Turn on for debug logs")
	rootCmd.AddCommand(cmdExportModel)

	var cmdLint = &cobra.Command{
		Use:   "lint",
		Short: "Evaluate Mendix model against policies. Requires the model to be exported first",
		Long:  "The model is evaluated against a set of policies. The policies are defined in OPA rego files. The output is a list of checked policies and their outcome.",
		Run: func(cmd *cobra.Command, args []string) {
			policiesDirectory, _ := cmd.Flags().GetString("policies")
			modelDirectory, _ := cmd.Flags().GetString("modelsource")
			xunitReport, _ := cmd.Flags().GetString("xunit-report")
			verbose, _ := cmd.Flags().GetBool("verbose")

			log := logrus.New()
			if verbose {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel)
			}

			lint.SetLogger(log)
			err := lint.EvalAll(policiesDirectory, modelDirectory, xunitReport)
			if err != nil {
				log.Errorf("lint failed: %s", err)
				os.Exit(1)
			}
		},
	}

	cmdLint.Flags().StringP("policies", "p", "policies", "Path to directory with policies")
	cmdLint.Flags().StringP("modelsource", "m", "modelsource", "Path to directory with exported model")
	cmdLint.Flags().StringP("xunit-report", "x", "", "Path to output file for xunit report. If not provided, no xunit report will be generated")
	cmdLint.Flags().Bool("verbose", false, "Turn on for debug logs")
	rootCmd.AddCommand(cmdLint)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
