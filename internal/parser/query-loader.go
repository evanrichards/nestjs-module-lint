package parser

import (
	_ "embed"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var typescriptLang = typescript.GetLanguage()

//go:embed module-imports.query
var moduleImportsQuery string

//go:embed import-paths.query
var importPathsQuery string

//go:embed module-exports.query
var moduleExportsQuery string

//go:embed module-provider-controller.query
var moduleProviderControllerQuery string

func queryFromString(queryContent string) (*sitter.Query, error) {
	return sitter.NewQuery([]byte(queryContent), typescriptLang)
}

func LoadModuleImportQuery() (*sitter.Query, error) {
	return queryFromString(moduleImportsQuery)
}

func LoadImportPathQuery() (*sitter.Query, error) {
	return queryFromString(importPathsQuery)
}

func LoadModuleExportQuery() (*sitter.Query, error) {
	return queryFromString(moduleExportsQuery)
}

func LoadModuleProviderControllerQuery() (*sitter.Query, error) {
	return queryFromString(moduleProviderControllerQuery)
}