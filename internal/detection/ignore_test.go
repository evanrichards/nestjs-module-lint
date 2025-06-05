package detection_test

import (
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/detection"
)

func TestIgnoreDetector_ParseIgnoreComments(t *testing.T) {
	detector := detection.NewIgnoreDetector()

	tests := []struct {
		name                   string
		sourceCode             string
		expectedFileIgnored    bool
		expectedIgnoredModules map[string]bool
		expectedIgnoredLines   map[int]bool
	}{
		{
			name: "file-level ignore",
			sourceCode: `// nestjs-module-lint-disable-file
import { Module } from '@nestjs/common';
import { UnusedModule } from './unused';`,
			expectedFileIgnored:    true,
			expectedIgnoredModules: map[string]bool{},
			expectedIgnoredLines:   map[int]bool{},
		},
		{
			name: "line-level ignore with module name",
			sourceCode: `import { Module } from '@nestjs/common';
import { UsedModule } from './used';
import { UnusedModule } from './unused';
@Module({
  imports: [
    UsedModule,
    UnusedModule, // nestjs-module-lint-disable-line
  ],
})`,
			expectedFileIgnored: false,
			expectedIgnoredModules: map[string]bool{
				"UnusedModule": true,
			},
			expectedIgnoredLines: map[int]bool{
				7: true,
			},
		},
		{
			name: "multiple line-level ignores",
			sourceCode: `Module1, // nestjs-module-lint-disable-line
import { Module2 } from './module2';
Module3, // nestjs-module-lint-disable-line`,
			expectedFileIgnored: false,
			expectedIgnoredModules: map[string]bool{
				"Module1": true,
				"Module3": true,
			},
			expectedIgnoredLines: map[int]bool{
				1: true,
				3: true,
			},
		},
		{
			name:                   "no ignore comments",
			sourceCode:             `import { Module } from '@nestjs/common';`,
			expectedFileIgnored:    false,
			expectedIgnoredModules: map[string]bool{},
			expectedIgnoredLines:   map[int]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := detector.ParseIgnoreComments([]byte(tt.sourceCode))

			if info.FileIgnored != tt.expectedFileIgnored {
				t.Errorf("FileIgnored = %v, want %v", info.FileIgnored, tt.expectedFileIgnored)
			}

			if len(info.IgnoredModules) != len(tt.expectedIgnoredModules) {
				t.Errorf("IgnoredModules count = %d, want %d", len(info.IgnoredModules), len(tt.expectedIgnoredModules))
			}

			for module, expected := range tt.expectedIgnoredModules {
				if info.IgnoredModules[module] != expected {
					t.Errorf("IgnoredModules[%s] = %v, want %v", module, info.IgnoredModules[module], expected)
				}
			}

			if len(info.IgnoredLines) != len(tt.expectedIgnoredLines) {
				t.Errorf("IgnoredLines count = %d, want %d", len(info.IgnoredLines), len(tt.expectedIgnoredLines))
			}
		})
	}
}

func TestIgnoreDetector_ShouldIgnoreFile(t *testing.T) {
	detector := detection.NewIgnoreDetector()

	tests := []struct {
		name       string
		sourceCode string
		expected   bool
	}{
		{
			name:       "file with ignore comment",
			sourceCode: "// nestjs-module-lint-disable-file\nimport { Module } from '@nestjs/common';",
			expected:   true,
		},
		{
			name:       "file without ignore comment",
			sourceCode: "import { Module } from '@nestjs/common';",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.ShouldIgnoreFile([]byte(tt.sourceCode))
			if result != tt.expected {
				t.Errorf("ShouldIgnoreFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIgnoreDetector_GetNonIgnoredImports(t *testing.T) {
	detector := detection.NewIgnoreDetector()

	sourceCode := []byte(`import { Module } from '@nestjs/common';
Module1, // nestjs-module-lint-disable-line
import { Module2 } from './module2';
Module3, // nestjs-module-lint-disable-line`)

	imports := []string{"Module1", "Module2", "Module3"}

	nonIgnored := detector.GetNonIgnoredImports(imports, sourceCode)

	if len(nonIgnored) != 1 {
		t.Errorf("Expected 1 non-ignored import, got %d", len(nonIgnored))
	}

	if nonIgnored[0] != "Module2" {
		t.Errorf("Expected Module2 to be non-ignored, got %s", nonIgnored[0])
	}
}
