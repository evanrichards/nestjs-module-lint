/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/evanrichards/nestjs-module-lint/internal/app"
	"github.com/spf13/cobra"
)

// importLintCmd represents the importLint command
var importLintCmd = &cobra.Command{
	Use:   "import-lint",
	Short: "Analyze NestJS modules for unused imports",
	Long: `Analyze NestJS modules for unused imports in @Module() decorators.

Exit codes:
  0 - No unused imports found (or --exit-zero flag used)
  1 - Unused imports found
  2 - Execution error (invalid path, parsing error, etc.)

Examples:
  # Basic usage
  nestjs-module-lint import-lint src/

  # Automatically fix unused imports
  nestjs-module-lint import-lint --fix src/

  # CI/CD usage with clear pass/fail
  nestjs-module-lint import-lint --check src/

  # Report only mode (always exit 0)
  nestjs-module-lint import-lint --exit-zero --quiet src/`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Handle fix mode separately
		if fixMode {
			for _, arg := range args {
				// Validate argument
				if strings.TrimSpace(arg) == "" {
					fmt.Fprintf(os.Stderr, "Error: empty path provided\n")
					os.Exit(2)
				}

				err := app.FixWorkflow(arg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fixing '%s': %v\n", arg, err)
					os.Exit(2)
				}
			}
			return
		}

		// Normal analysis mode
		var allReports []*app.ModuleReport

		for _, arg := range args {
			// Validate argument
			if strings.TrimSpace(arg) == "" {
				fmt.Fprintf(os.Stderr, "Error: empty path provided\n")
				os.Exit(2)
			}

			reports, err := app.AnalyzePath(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing '%s': %v\n", arg, err)
				os.Exit(2) // Exit code 2 for execution errors
			}
			allReports = append(allReports, reports...)
		}

		// Output results based on format
		if ofJson {
			d, _ := json.Marshal(allReports)
			fmt.Println(string(d))
		} else if !quiet {
			// Text output (default)
			for _, report := range allReports {
				fmt.Println(app.PrettyPrintModuleReport(report))
			}

			if checkMode {
				if len(allReports) > 0 {
					fmt.Printf("✗ Found %d modules with unused imports\n", len(allReports))
				} else {
					fmt.Println("✓ No unused imports found")
				}
			} else {
				fmt.Printf("Total number of modules with unused imports: %d\n", len(allReports))
			}
		}

		// Determine exit code
		if len(allReports) > 0 && !exitZero {
			os.Exit(1) // Exit code 1 for linting failures
		}
	},
}

var ofJson bool
var ofText bool
var exitZero bool
var checkMode bool
var quiet bool
var fixMode bool

func init() {
	rootCmd.AddCommand(importLintCmd)

	// Output format flags
	importLintCmd.Flags().BoolVar(&ofJson, "json", false, "Output in JSON format")
	importLintCmd.Flags().BoolVar(&ofText, "text", false, "Output in text format")

	// CI/CD flags
	importLintCmd.Flags().BoolVar(&checkMode, "check", false, "Check mode with pass/fail output (good for CI)")
	importLintCmd.Flags().BoolVar(&exitZero, "exit-zero", false, "Exit with code 0 even when issues are found")
	importLintCmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress output (useful with --exit-zero)")

	// Fix flags
	importLintCmd.Flags().BoolVar(&fixMode, "fix", false, "Automatically remove unused imports")

	importLintCmd.MarkFlagsMutuallyExclusive("json", "text")
	importLintCmd.MarkFlagsMutuallyExclusive("fix", "json")
	importLintCmd.MarkFlagsMutuallyExclusive("fix", "check")
}
