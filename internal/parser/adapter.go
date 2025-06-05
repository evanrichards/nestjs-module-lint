package parser

import (
	"context"
	"os"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
	sitter "github.com/smacker/go-tree-sitter"
)

// ParserAdapter adapts the existing parser functions to implement the ModuleParser interface
type ParserAdapter struct {
	lang *sitter.Language
}

// NewParserAdapter creates a new parser adapter
func NewParserAdapter(lang *sitter.Language) *ParserAdapter {
	return &ParserAdapter{
		lang: lang,
	}
}

// ParseModuleInfo implements the ModuleParser interface
func (p *ParserAdapter) ParseModuleInfo(filePath string) (*analysis.ModuleInfo, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	// Get imports by module
	importsByModule, err := ParseModuleImports(tree, sourceCode)
	if err != nil {
		return nil, err
	}

	// Get exports by module
	exportsByModule, err := ParseModuleExports(tree, sourceCode)
	if err != nil {
		return nil, err
	}

	// Get providers by module
	providersByModule, err := ParseModuleProviders(tree, sourceCode)
	if err != nil {
		return nil, err
	}

	// For simplicity, take the first module found
	// TODO: This could be improved to handle multiple modules per file
	for moduleName := range importsByModule {
		return &analysis.ModuleInfo{
			Name:      moduleName,
			FilePath:  filePath,
			Imports:   importsByModule[moduleName],
			Exports:   exportsByModule[moduleName],
			Providers: providersByModule[moduleName],
		}, nil
	}

	// No modules found
	return nil, nil
}

// GetImportsByModule implements the ModuleParser interface
func (p *ParserAdapter) GetImportsByModule(filePath string) (map[string][]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseModuleImports(tree, sourceCode)
}

// GetExportsByModule implements the ModuleParser interface
func (p *ParserAdapter) GetExportsByModule(filePath string) (map[string][]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseModuleExports(tree, sourceCode)
}

// GetProvidersByModule implements the ModuleParser interface
func (p *ParserAdapter) GetProvidersByModule(filePath string) (map[string][]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseModuleProviders(tree, sourceCode)
}

// GetImportPaths implements the ModuleParser interface
func (p *ParserAdapter) GetImportPaths(filePath string) (map[string]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseImportPaths(tree, sourceCode)
}
