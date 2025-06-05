package fixing_test

import (
	"strings"
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/fixing"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func TestFixer_FixUnusedImports(t *testing.T) {
	lang := typescript.GetLanguage()
	fixer := fixing.NewFixer(lang)

	tests := []struct {
		name           string
		sourceCode     string
		unusedModules  []string
		expectedResult string
	}{
		{
			name: "remove single unused import",
			sourceCode: `import { Module } from "@nestjs/common";
import { UnusedModule } from "./unused.module";
import { UsedModule } from "./used.module";

@Module({
  imports: [UnusedModule, UsedModule],
  providers: [],
})
export class AppModule {}`,
			unusedModules: []string{"UnusedModule"},
			expectedResult: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "remove multiple unused imports",
			sourceCode: `import { Module } from "@nestjs/common";
import { Unused1 } from "./unused1";
import { Unused2 } from "./unused2";
import { UsedModule } from "./used.module";

@Module({
  imports: [Unused1, UsedModule, Unused2],
  providers: [],
})
export class AppModule {}`,
			unusedModules: []string{"Unused1", "Unused2"},
			expectedResult: `import { Module } from "@nestjs/common";
import { UsedModule } from "./used.module";

@Module({
  imports: [UsedModule],
  providers: [],
})
export class AppModule {}`,
		},
		{
			name: "handle multiline imports array",
			sourceCode: `import { Module } from "@nestjs/common";
import { UnusedModule } from "./unused.module";
import { UsedModule } from "./used.module";

@Module({
  imports: [
    UnusedModule,
    UsedModule,
  ],
  providers: [],
})
export class AppModule {}`,
			unusedModules: []string{"UnusedModule"},
			expectedResult: `import { Module } from "@nestjs/common";
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
			name:           "no modules to remove",
			sourceCode:     `import { Module } from "@nestjs/common";`,
			unusedModules:  []string{},
			expectedResult: `import { Module } from "@nestjs/common";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fixer.FixUnusedImports([]byte(tt.sourceCode), tt.unusedModules)
			if err != nil {
				t.Fatalf("FixUnusedImports failed: %v", err)
			}

			resultStr := string(result)
			// Normalize line endings for comparison
			resultStr = strings.ReplaceAll(resultStr, "\r\n", "\n")
			expectedStr := strings.ReplaceAll(tt.expectedResult, "\r\n", "\n")

			if strings.TrimSpace(resultStr) != strings.TrimSpace(expectedStr) {
				t.Errorf("Result mismatch\nGot:\n%s\n\nExpected:\n%s", resultStr, expectedStr)
			}
		})
	}
}

func TestFixer_FixUnusedImports_InvalidSyntax(t *testing.T) {
	lang := typescript.GetLanguage()
	fixer := fixing.NewFixer(lang)

	// Test with invalid TypeScript syntax
	invalidCode := `import { Module from "@nestjs/common"; // missing closing brace`

	_, err := fixer.FixUnusedImports([]byte(invalidCode), []string{"Module"})
	if err == nil {
		t.Error("Expected error for invalid TypeScript syntax")
	}
}
