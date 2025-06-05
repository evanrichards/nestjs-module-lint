package parser_test

import (
	"context"
	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"testing"
)

func TestGetExportsByModuleFromFile(t *testing.T) {
	// Example TypeScript source to parse
	sourceCode := `
import { Module } from "@nestjs/common";
import { SomeExport } from "./some-export";
@Module({
  exports: [SomeExport],
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
	exportsByModule, err := parser.GetExportsByModuleFromFile(node, []byte(sourceCode))
	if err != nil {
		t.Fatalf("Failed to get exports by module: %v", err)
	}

	// Verify expected output
	expectedExports := map[string][]string{
		"AppModule": {"SomeExport"},
	}
	for moduleName, exports := range expectedExports {
		if gotExports, ok := exportsByModule[moduleName]; !ok || len(gotExports) != len(exports) {
			t.Errorf("Expected exports %v for module %v, got %v", exports, moduleName, gotExports)
		}
	}
}
