package parser_test

import (
	"context"
	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"testing"
)

func TestParseModuleProviders(t *testing.T) {
	// Example TypeScript source to parse
	sourceCode := `
import { Module } from "@nestjs/common";
import { SomeService, RestSomeService } from "./some-import";
@Module({
  providers: [SomeService],
  controllers: [RestSomeService],
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
	providersControllerByModule, err := parser.ParseModuleProviders(node, []byte(sourceCode))
	if err != nil {
		t.Fatalf("Failed to get provider controllers by module: %v", err)
	}

	// Verify expected output
	expected := map[string][]string{
		"AppModule": {"SomeService", "RestSomeService"},
	}
	for moduleName, providerOrController := range expected {
		if gotProviderOrController, ok := providersControllerByModule[moduleName]; !ok || len(gotProviderOrController) != len(providerOrController) {
			t.Errorf("Expected providers or controllers %v for module %v, got %v", providerOrController, moduleName, gotProviderOrController)
		}
	}
}
