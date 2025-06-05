package reporting

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
)

// OutputFormat defines the format for output
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// Formatter handles formatting of analysis results
type Formatter struct{}

// NewFormatter creates a new result formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format formats the analysis results according to the specified format
func (f *Formatter) Format(results []*analysis.ModuleAnalysisResult, format OutputFormat) (string, error) {
	switch format {
	case FormatJSON:
		return f.formatJSON(results)
	case FormatText:
		return f.formatText(results), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatJSON formats results as JSON
func (f *Formatter) formatJSON(results []*analysis.ModuleAnalysisResult) (string, error) {
	data, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatText formats results as human-readable text
func (f *Formatter) formatText(results []*analysis.ModuleAnalysisResult) string {
	if len(results) == 0 {
		return "No unused imports found."
	}

	var builder strings.Builder

	for i, result := range results {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(f.formatModuleResult(result))
	}

	builder.WriteString(fmt.Sprintf("\nTotal number of modules with unused imports: %d\n", len(results)))

	return builder.String()
}

// formatModuleResult formats a single module result
func (f *Formatter) formatModuleResult(result *analysis.ModuleAnalysisResult) string {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("Module: %s\n", result.ModuleName))
	builder.WriteString(fmt.Sprintf("Path: %s\n", result.FilePath))

	if len(result.UnusedImports) > 0 {
		builder.WriteString("Unused Imports:\n")
		for _, imp := range result.UnusedImports {
			builder.WriteString(fmt.Sprintf("\t%s\n", imp))
		}
	}

	if len(result.IgnoredImports) > 0 {
		builder.WriteString("Ignored Imports:\n")
		for _, imp := range result.IgnoredImports {
			builder.WriteString(fmt.Sprintf("\t%s (ignored)\n", imp))
		}
	}

	if len(result.ReExportedImports) > 0 {
		builder.WriteString("Re-exported Imports:\n")
		for _, imp := range result.ReExportedImports {
			builder.WriteString(fmt.Sprintf("\t%s (re-exported)\n", imp))
		}
	}

	return builder.String()
}

// GetSummary returns a summary message for check mode
func (f *Formatter) GetSummary(results []*analysis.ModuleAnalysisResult, checkMode bool) string {
	if len(results) == 0 {
		if checkMode {
			return "✓ No unused imports found"
		}
		return "No unused imports found."
	}

	if checkMode {
		return fmt.Sprintf("✗ Found %d modules with unused imports", len(results))
	}

	return fmt.Sprintf("Total number of modules with unused imports: %d", len(results))
}
