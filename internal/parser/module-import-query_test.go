package parser_test

import (
	"context"
	"github.com/loop-payments/nestjs-module-lint/internal/parser"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"testing"
)

func TestGetImportsByModuleFromFile(t *testing.T) {
	// Example TypeScript source to parse
	sourceCode := `
import { Module } from "@nestjs/common";
import { SomeImport } from "./some-import";
@Module({
  imports: [SomeImport],
})
export class AppModule {}
`

	// Parse the source code into an AST
	lang := typescript.GetLanguage()
	node, err := sitter.ParseCtx(context.Background(), []byte(sourceCode), lang)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}

	// Call the function under test
	importsByModule, err := parser.GetImportsByModuleFromFile(node, []byte(sourceCode))
	if err != nil {
		t.Fatalf("Failed to get imports by module: %v", err)
	}

	// Verify expected output
	expectedImports := map[string][]string{
		"AppModule": {"SomeImport"},
	}
	for moduleName, imports := range expectedImports {
		if gotImports, ok := importsByModule[moduleName]; !ok || len(gotImports) != len(imports) {
			t.Errorf("Expected imports %v for module %v, got %v", imports, moduleName, gotImports)
		}
	}
}
