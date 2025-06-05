package detection_test

import (
	"reflect"
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/detection"
)

func TestReExportDetector_GetReExportedModules(t *testing.T) {
	detector := detection.NewReExportDetector()

	tests := []struct {
		name     string
		imports  []string
		exports  []string
		expected []string
	}{
		{
			name:     "no exports - no re-exports",
			imports:  []string{"ModuleA", "ModuleB"},
			exports:  []string{},
			expected: []string{},
		},
		{
			name:     "all imports are re-exported",
			imports:  []string{"ModuleA", "ModuleB"},
			exports:  []string{"ModuleA", "ModuleB"},
			expected: []string{"ModuleA", "ModuleB"},
		},
		{
			name:     "partial re-export",
			imports:  []string{"ModuleA", "ModuleB", "ModuleC"},
			exports:  []string{"ModuleA", "ModuleC"},
			expected: []string{"ModuleA", "ModuleC"},
		},
		{
			name:     "exports include modules not in imports",
			imports:  []string{"ModuleA"},
			exports:  []string{"ModuleA", "ModuleB", "LocalModule"},
			expected: []string{"ModuleA"},
		},
		{
			name:     "no imports",
			imports:  []string{},
			exports:  []string{"ModuleA", "ModuleB"},
			expected: nil, // GetReExportedModules returns nil for empty imports, not empty slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.GetReExportedModules(tt.imports, tt.exports)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetReExportedModules() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReExportDetector_GetNonReExportedImports(t *testing.T) {
	detector := detection.NewReExportDetector()

	tests := []struct {
		name     string
		imports  []string
		exports  []string
		expected []string
	}{
		{
			name:     "no exports - all imports remain",
			imports:  []string{"ModuleA", "ModuleB"},
			exports:  []string{},
			expected: []string{"ModuleA", "ModuleB"},
		},
		{
			name:     "all imports are re-exported - none remain",
			imports:  []string{"ModuleA", "ModuleB"},
			exports:  []string{"ModuleA", "ModuleB"},
			expected: nil, // Function returns nil when no items remain after filtering
		},
		{
			name:     "partial re-export",
			imports:  []string{"ModuleA", "ModuleB", "ModuleC"},
			exports:  []string{"ModuleA", "ModuleC"},
			expected: []string{"ModuleB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.GetNonReExportedImports(tt.imports, tt.exports)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetNonReExportedImports() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReExportDetector_IsReExported(t *testing.T) {
	detector := detection.NewReExportDetector()

	exports := []string{"ModuleA", "ModuleB", "ModuleC"}

	tests := []struct {
		moduleName string
		expected   bool
	}{
		{"ModuleA", true},
		{"ModuleB", true},
		{"ModuleC", true},
		{"ModuleD", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.moduleName, func(t *testing.T) {
			result := detector.IsReExported(tt.moduleName, exports)
			if result != tt.expected {
				t.Errorf("IsReExported(%s) = %v, want %v", tt.moduleName, result, tt.expected)
			}
		})
	}
}
