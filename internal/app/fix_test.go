package app_test

import (
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/app"
)

func TestFixUnusedImports(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		toRemove []string
		expected string
	}{
		{
			name: "remove single unused import from inline array",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";
import { UnusedModule } from "./unused.module";

@Module({
  imports: [UsedModule, UnusedModule],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove multiple unused imports from inline array",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";
import { UnusedA } from "./unused-a.module";
import { UnusedB } from "./unused-b.module";

@Module({
  imports: [UsedModule, UnusedA, UnusedB],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedA", "UnusedB"},
			expected: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove unused import from multiline array",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";
import { UnusedModule } from "./unused.module";

@Module({
  imports: [
    UsedModule,
    UnusedModule,
  ],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [
    UsedModule,
  ],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove unused import from multiline array with trailing comma",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";
import { UnusedModule } from "./unused.module";

@Module({
  imports: [
    UsedModule,
    UnusedModule,
  ],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [
    UsedModule,
  ],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove all unused imports leaving empty array",
			input: `import { Module } from "@nestjs/common";
import { UnusedA } from "./unused-a.module";
import { UnusedB } from "./unused-b.module";

@Module({
  imports: [UnusedA, UnusedB],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedA", "UnusedB"},
			expected: `import { Module } from "@nestjs/common";

@Module({
  imports: [],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove first import from inline array",
			input: `import { Module } from "@nestjs/common";
import { UnusedModule } from "./unused.module";
import { UsedModule } from "./used.module";

@Module({
  imports: [UnusedModule, UsedModule],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove middle import from inline array",
			input: `import { Module } from "@nestjs/common";
import { UsedA } from "./used-a.module";
import { UnusedModule } from "./unused.module";
import { UsedB } from "./used-b.module";

@Module({
  imports: [UsedA, UnusedModule, UsedB],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { UsedA } from "./used-a.module";
import { UsedB } from "./used-b.module";

@Module({
  imports: [UsedA, UsedB],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "handle named imports with aliasing",
			input: `import { Module } from "@nestjs/common";
import { SomeModule as AliasedModule } from "./some.module";
import { UnusedModule } from "./unused.module";

@Module({
  imports: [AliasedModule, UnusedModule],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { SomeModule as AliasedModule } from "./some.module";

@Module({
  imports: [AliasedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "handle default imports",
			input: `import { Module } from "@nestjs/common";
import UsedModule from "./used.module";
import UnusedModule from "./unused.module";

@Module({
  imports: [UsedModule, UnusedModule],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import UsedModule from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "handle complex spacing and comments",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";
import { UnusedModule } from "./unused.module"; // This module is not used

@Module({
  imports: [
    UsedModule, // This one is used
    UnusedModule, // This one should be removed
  ],
  providers: [],
})
export class AppModule {}`,
			toRemove: []string{"UnusedModule"},
			expected: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [
    UsedModule, // This one is used
  ],
  providers: [],
})
export class AppModule {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := app.FixUnusedImports([]byte(tt.input), tt.toRemove)
			if err != nil {
				t.Fatalf("FixUnusedImports failed: %v", err)
			}

			if string(result) != tt.expected {
				t.Errorf("FixUnusedImports result mismatch\nGot:\n%s\n\nExpected:\n%s", string(result), tt.expected)
			}
		})
	}
}

func TestFixUnusedImportsEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		toRemove    []string
		expectError bool
	}{
		{
			name: "no modules to remove",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
			toRemove:    []string{},
			expectError: false,
		},
		{
			name: "module not found in imports array",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
			toRemove:    []string{"NonExistentModule"},
			expectError: false, // Should not error, just skip
		},
		{
			name: "malformed TypeScript",
			input: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module"

@Module({
  imports: [UsedModule
  providers: [],
})
export class AppModule {}`,
			toRemove:    []string{"UsedModule"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := app.FixUnusedImports([]byte(tt.input), tt.toRemove)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
