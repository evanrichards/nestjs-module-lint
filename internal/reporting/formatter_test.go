package reporting_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
	"github.com/evanrichards/nestjs-module-lint/internal/reporting"
)

func TestFormatter_Format_JSON(t *testing.T) {
	formatter := reporting.NewFormatter()

	results := []*analysis.ModuleAnalysisResult{
		{
			ModuleName:    "AppModule",
			FilePath:      "src/app.module.ts",
			UnusedImports: []string{"UnusedModule1", "UnusedModule2"},
		},
		{
			ModuleName:    "UserModule",
			FilePath:      "src/user/user.module.ts",
			UnusedImports: []string{"UnusedService"},
		},
	}

	output, err := formatter.Format(results, reporting.FormatJSON)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed []analysis.ModuleAnalysisResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}

	// Verify content
	if len(parsed) != 2 {
		t.Errorf("Expected 2 results in JSON, got %d", len(parsed))
	}
}

func TestFormatter_Format_Text(t *testing.T) {
	formatter := reporting.NewFormatter()

	results := []*analysis.ModuleAnalysisResult{
		{
			ModuleName:        "AppModule",
			FilePath:          "src/app.module.ts",
			UnusedImports:     []string{"UnusedModule1", "UnusedModule2"},
			IgnoredImports:    []string{"IgnoredModule"},
			ReExportedImports: []string{"ReExportedModule"},
		},
	}

	output, err := formatter.Format(results, reporting.FormatText)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify output contains expected content
	expectedStrings := []string{
		"Module: AppModule",
		"Path: src/app.module.ts",
		"Unused Imports:",
		"UnusedModule1",
		"UnusedModule2",
		"Ignored Imports:",
		"IgnoredModule (ignored)",
		"Re-exported Imports:",
		"ReExportedModule (re-exported)",
		"Total number of modules with unused imports: 1",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

func TestFormatter_Format_EmptyResults(t *testing.T) {
	formatter := reporting.NewFormatter()

	results := []*analysis.ModuleAnalysisResult{}

	// Test text format
	output, err := formatter.Format(results, reporting.FormatText)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if output != "No unused imports found." {
		t.Errorf("Expected 'No unused imports found.', got '%s'", output)
	}

	// Test JSON format
	output, err = formatter.Format(results, reporting.FormatJSON)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if output != "[]" {
		t.Errorf("Expected empty JSON array, got '%s'", output)
	}
}

func TestFormatter_Format_InvalidFormat(t *testing.T) {
	formatter := reporting.NewFormatter()

	results := []*analysis.ModuleAnalysisResult{}

	_, err := formatter.Format(results, "invalid-format")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestFormatter_GetSummary(t *testing.T) {
	formatter := reporting.NewFormatter()

	tests := []struct {
		name      string
		results   []*analysis.ModuleAnalysisResult
		checkMode bool
		expected  string
	}{
		{
			name:      "empty results with check mode",
			results:   []*analysis.ModuleAnalysisResult{},
			checkMode: true,
			expected:  "✓ No unused imports found",
		},
		{
			name:      "empty results without check mode",
			results:   []*analysis.ModuleAnalysisResult{},
			checkMode: false,
			expected:  "No unused imports found.",
		},
		{
			name: "results with check mode",
			results: []*analysis.ModuleAnalysisResult{
				{ModuleName: "Module1", UnusedImports: []string{"Import1"}},
				{ModuleName: "Module2", UnusedImports: []string{"Import2"}},
			},
			checkMode: true,
			expected:  "✗ Found 2 modules with unused imports",
		},
		{
			name: "results without check mode",
			results: []*analysis.ModuleAnalysisResult{
				{ModuleName: "Module1", UnusedImports: []string{"Import1"}},
			},
			checkMode: false,
			expected:  "Total number of modules with unused imports: 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := formatter.GetSummary(tt.results, tt.checkMode)
			if summary != tt.expected {
				t.Errorf("GetSummary() = %s, want %s", summary, tt.expected)
			}
		})
	}
}
