package parser

import (
	_ "embed"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var (
	typescriptLang = typescript.GetLanguage()
	
	// Cached compiled queries
	moduleImportQueryCache     *sitter.Query
	importPathQueryCache       *sitter.Query
	moduleExportQueryCache     *sitter.Query
	moduleProviderQueryCache   *sitter.Query
	
	// Sync guards for one-time initialization
	moduleImportQueryOnce     sync.Once
	importPathQueryOnce       sync.Once
	moduleExportQueryOnce     sync.Once
	moduleProviderQueryOnce   sync.Once
)

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
	var err error
	moduleImportQueryOnce.Do(func() {
		moduleImportQueryCache, err = queryFromString(moduleImportsQuery)
	})
	return moduleImportQueryCache, err
}

func LoadImportPathQuery() (*sitter.Query, error) {
	var err error
	importPathQueryOnce.Do(func() {
		importPathQueryCache, err = queryFromString(importPathsQuery)
	})
	return importPathQueryCache, err
}

func LoadModuleExportQuery() (*sitter.Query, error) {
	var err error
	moduleExportQueryOnce.Do(func() {
		moduleExportQueryCache, err = queryFromString(moduleExportsQuery)
	})
	return moduleExportQueryCache, err
}

func LoadModuleProviderControllerQuery() (*sitter.Query, error) {
	var err error
	moduleProviderQueryOnce.Do(func() {
		moduleProviderQueryCache, err = queryFromString(moduleProviderControllerQuery)
	})
	return moduleProviderQueryCache, err
}
