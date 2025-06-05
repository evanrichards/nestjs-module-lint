package app_test

import (
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/app"
)

func TestParseIgnoreComments(t *testing.T) {
	tests := []struct {
		name               string
		sourceCode         string
		expectedFileIgnored bool
		expectedIgnoredModules map[string]bool
		expectedIgnoredLines map[int]bool
	}{
		{
			name: "file-level ignore",
			sourceCode: `// nestjs-module-lint-disable-file
import { Module } from '@nestjs/common';
import { UnusedModule } from './unused.module';

@Module({
  imports: [UnusedModule],
})
export class TestModule {}`,
			expectedFileIgnored: true,
			expectedIgnoredModules: map[string]bool{},
			expectedIgnoredLines: map[int]bool{},
		},
		{
			name: "line-level ignore with module name",
			sourceCode: `import { Module } from '@nestjs/common';
import { UsedModule } from './used.module';
import { UnusedModuleA } from './unused-a.module';
import { UnusedModuleB } from './unused-b.module';

@Module({
  imports: [
    UsedModule,
    UnusedModuleA, // nestjs-module-lint-disable-line
    UnusedModuleB,
  ],
})
export class TestModule {}`,
			expectedFileIgnored: false,
			expectedIgnoredModules: map[string]bool{"UnusedModuleA": true},
			expectedIgnoredLines: map[int]bool{9: true},
		},
		{
			name: "multiple line-level ignores",
			sourceCode: `import { Module } from '@nestjs/common';
import { UnusedModuleA } from './unused-a.module';
import { UnusedModuleB } from './unused-b.module';
import { UnusedModuleC } from './unused-c.module';

@Module({
  imports: [
    UnusedModuleA, // nestjs-module-lint-disable-line
    UnusedModuleB,
    UnusedModuleC, // nestjs-module-lint-disable-line
  ],
})
export class TestModule {}`,
			expectedFileIgnored: false,
			expectedIgnoredModules: map[string]bool{
				"UnusedModuleA": true,
				"UnusedModuleC": true,
			},
			expectedIgnoredLines: map[int]bool{8: true, 10: true},
		},
		{
			name: "no ignore comments",
			sourceCode: `import { Module } from '@nestjs/common';
import { UnusedModule } from './unused.module';

@Module({
  imports: [UnusedModule],
})
export class TestModule {}`,
			expectedFileIgnored: false,
			expectedIgnoredModules: map[string]bool{},
			expectedIgnoredLines: map[int]bool{},
		},
		{
			name: "ignore with spacing variations",
			sourceCode: `import { Module } from '@nestjs/common';
import { UnusedModuleA } from './unused-a.module';
import { UnusedModuleB } from './unused-b.module';

@Module({
  imports: [
    UnusedModuleA,//nestjs-module-lint-disable-line
    UnusedModuleB, //   nestjs-module-lint-disable-line   
  ],
})
export class TestModule {}`,
			expectedFileIgnored: false,
			expectedIgnoredModules: map[string]bool{
				"UnusedModuleA": true,
				"UnusedModuleB": true,
			},
			expectedIgnoredLines: map[int]bool{7: true, 8: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := app.ParseIgnoreComments([]byte(tt.sourceCode))
			
			if info.FileIgnored != tt.expectedFileIgnored {
				t.Errorf("Expected FileIgnored %v, got %v", tt.expectedFileIgnored, info.FileIgnored)
			}
			
			// Check ignored modules
			for module, expected := range tt.expectedIgnoredModules {
				if info.ShouldIgnoreModule(module) != expected {
					t.Errorf("Expected module %s to be ignored: %v, got %v", module, expected, info.ShouldIgnoreModule(module))
				}
			}
			
			// Check ignored lines
			for line, expected := range tt.expectedIgnoredLines {
				if info.IgnoredLines[line] != expected {
					t.Errorf("Expected line %d to be ignored: %v, got %v", line, expected, info.IgnoredLines[line])
				}
			}
		})
	}
}

func TestShouldIgnoreModule(t *testing.T) {
	tests := []struct {
		name       string
		ignoreInfo *app.IgnoreInfo
		moduleName string
		expected   bool
	}{
		{
			name: "file ignored - should ignore any module",
			ignoreInfo: &app.IgnoreInfo{
				FileIgnored:    true,
				IgnoredLines:   map[int]bool{},
				IgnoredModules: map[string]bool{},
			},
			moduleName: "AnyModule",
			expected:   true,
		},
		{
			name: "module specifically ignored",
			ignoreInfo: &app.IgnoreInfo{
				FileIgnored:    false,
				IgnoredLines:   map[int]bool{},
				IgnoredModules: map[string]bool{"IgnoredModule": true},
			},
			moduleName: "IgnoredModule",
			expected:   true,
		},
		{
			name: "module not ignored",
			ignoreInfo: &app.IgnoreInfo{
				FileIgnored:    false,
				IgnoredLines:   map[int]bool{},
				IgnoredModules: map[string]bool{"IgnoredModule": true},
			},
			moduleName: "NotIgnoredModule",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ignoreInfo.ShouldIgnoreModule(tt.moduleName)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}