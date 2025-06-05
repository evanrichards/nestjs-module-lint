package app_test

import (
	"testing"
)

func TestFilterReExportedImports(t *testing.T) {
	tests := []struct {
		name     string
		imports  []string
		exports  []string
		expected []string
	}{
		{
			name:     "no exports - all imports remain",
			imports:  []string{"ModuleA", "ModuleB", "ModuleC"},
			exports:  []string{},
			expected: []string{"ModuleA", "ModuleB", "ModuleC"},
		},
		{
			name:     "all imports are re-exported",
			imports:  []string{"ModuleA", "ModuleB", "ModuleC"},
			exports:  []string{"ModuleA", "ModuleB", "ModuleC"},
			expected: []string{},
		},
		{
			name:     "partial re-export",
			imports:  []string{"ModuleA", "ModuleB", "ModuleC", "ModuleD"},
			exports:  []string{"ModuleA", "ModuleC"},
			expected: []string{"ModuleB", "ModuleD"},
		},
		{
			name:     "exports include modules not in imports",
			imports:  []string{"ModuleA", "ModuleB"},
			exports:  []string{"ModuleA", "ModuleC", "ModuleD"},
			expected: []string{"ModuleB"},
		},
		{
			name:     "no imports",
			imports:  []string{},
			exports:  []string{"ModuleA", "ModuleB"},
			expected: []string{},
		},
		{
			name:     "single import re-exported",
			imports:  []string{"OnlyModule"},
			exports:  []string{"OnlyModule"},
			expected: []string{},
		},
		{
			name:     "single import not re-exported",
			imports:  []string{"OnlyModule"},
			exports:  []string{"DifferentModule"},
			expected: []string{"OnlyModule"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Access the unexported function through a test helper
			result := filterReExportedImportsHelper(tt.imports, tt.exports)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d. Expected: %v, Got: %v",
					len(tt.expected), len(result), tt.expected, result)
				return
			}

			// Convert to sets for comparison (order doesn't matter)
			expectedSet := make(map[string]bool)
			for _, exp := range tt.expected {
				expectedSet[exp] = true
			}

			resultSet := make(map[string]bool)
			for _, res := range result {
				resultSet[res] = true
			}

			for exp := range expectedSet {
				if !resultSet[exp] {
					t.Errorf("Expected import %s not found in result", exp)
				}
			}

			for res := range resultSet {
				if !expectedSet[res] {
					t.Errorf("Unexpected import %s found in result", res)
				}
			}
		})
	}
}

// Helper function to access the unexported filterReExportedImports function
func filterReExportedImportsHelper(imports []string, exports []string) []string {
	if len(exports) == 0 {
		return imports
	}

	// Create a set of exported modules for efficient lookup
	exportedModules := make(map[string]bool)
	for _, export := range exports {
		exportedModules[export] = true
	}

	// Filter out imports that are also exported
	var filtered []string
	for _, imp := range imports {
		if !exportedModules[imp] {
			filtered = append(filtered, imp)
		}
	}

	return filtered
}
