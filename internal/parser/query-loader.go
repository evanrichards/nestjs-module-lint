package parser

import (
	"os"
	"path/filepath"
	"runtime"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var typescriptLang = typescript.GetLanguage()

var filename string

func init() {
	_, _filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file path")
	}
	filename = filepath.Dir(_filename)
}

func queryForFile(queryName string) (*sitter.Query, error) {
	queryPath := filepath.Join(filename, queryName)
	query, err := os.ReadFile(queryPath)
	if err != nil {
		return nil, err
	}
	return sitter.NewQuery(query, typescriptLang)
}

func LoadModuleImportQuery() (*sitter.Query, error) {
	return queryForFile("module-imports.query")
}

func LoadImportPathQuery() (*sitter.Query, error) {
	return queryForFile("import-paths.query")
}

func LoadModuleExportQuery() (*sitter.Query, error) {
	return queryForFile("module-exports.query")
}
