/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/loop-payments/nestjs-module-lint/internal/app"
	"github.com/spf13/cobra"
)

// importLintCmd represents the importLint command
var importLintCmd = &cobra.Command{
	Use:   "importLint",
	Short: "Unused import linting",
	Long:  `This command will lint your project for unused module imports.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			reports, err := app.RunForDirRecursively(arg)
			if err != nil {
				panic(err)
			}
			if ofJson {
				d, _ := json.Marshal(reports)
				fmt.Println(string(d))
			}
			if ofText || (!ofJson && !ofText) {
				for _, report := range reports {
					fmt.Println(app.PrettyPrintModuleReport(report))
				}
				fmt.Printf("Total number of modules with unused imports: %d\n", len(reports))
			}
			if len(reports) > 0 {
				os.Exit(1)
			}
		}
	},
}

var ofJson bool
var ofText bool

func init() {
	rootCmd.AddCommand(importLintCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	importLintCmd.Flags().BoolVar(&ofJson, "json", false, "Output in JSON")
	importLintCmd.Flags().BoolVar(&ofText, "text", false, "Output in text")
	importLintCmd.MarkFlagsMutuallyExclusive("json", "text")
}
