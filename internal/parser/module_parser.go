package parser

import (
	"context"
	"os"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
	sitter "github.com/smacker/go-tree-sitter"
)

// ModuleParser implements the analysis.ModuleParser interface
type ModuleParser struct {
	lang *sitter.Language
}

// NewModuleParser creates a new module parser
func NewModuleParser(lang *sitter.Language) *ModuleParser {
	return &ModuleParser{
		lang: lang,
	}
}

// ParseModuleInfo parses basic module information from a file
func (p *ModuleParser) ParseModuleInfo(filePath string) (*analysis.ModuleInfo, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	n, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	// Get imports by module
	importsByModule, err := ParseModuleImports(n, sourceCode)
	if err != nil {
		return nil, err
	}

	// Get exports by module
	exportsByModule, err := ParseModuleExports(n, sourceCode)
	if err != nil {
		return nil, err
	}

	// Get providers by module
	providersByModule, err := ParseModuleProviders(n, sourceCode)
	if err != nil {
		return nil, err
	}

	// For simplicity, take the first module found
	// In practice, most files have one module
	for moduleName := range importsByModule {
		return &analysis.ModuleInfo{
			Name:      moduleName,
			FilePath:  filePath,
			Imports:   importsByModule[moduleName],
			Exports:   exportsByModule[moduleName],
			Providers: providersByModule[moduleName],
		}, nil
	}

	// If no modules found, return empty info
	return &analysis.ModuleInfo{
		Name:      "",
		FilePath:  filePath,
		Imports:   []string{},
		Exports:   []string{},
		Providers: []string{},
	}, nil
}

// GetImportsByModule returns imports grouped by module name
func (p *ModuleParser) GetImportsByModule(filePath string) (map[string][]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	n, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseModuleImports(n, sourceCode)
}

// GetExportsByModule returns exports grouped by module name
func (p *ModuleParser) GetExportsByModule(filePath string) (map[string][]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	n, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseModuleExports(n, sourceCode)
}

// GetProvidersByModule returns providers grouped by module name
func (p *ModuleParser) GetProvidersByModule(filePath string) (map[string][]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	n, err := sitter.ParseCtx(context.Background(), sourceCode, p.lang)
	if err != nil {
		return nil, err
	}

	return ParseModuleProviders(n, sourceCode)
}
