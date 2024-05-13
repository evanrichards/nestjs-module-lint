package parser

import (
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"path/filepath"
	"runtime"
)

var typescriptLang = typescript.GetLanguage()

func LoadModuleImportQuery() (*sitter.Query, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, os.ErrNotExist
	}
	queryPath := filepath.Join(filepath.Dir(filename), "module-imports.query")
	importsQuery, err := os.ReadFile(queryPath)
	if err != nil {
		return nil, err
	}
	return sitter.NewQuery([]byte(importsQuery), typescriptLang)
}

func LoadImportPathQuery() (*sitter.Query, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, os.ErrNotExist
	}
	queryPath := filepath.Join(filepath.Dir(filename), "import-paths.query")
	importsQuery, err := os.ReadFile(queryPath)
	if err != nil {
		return nil, err
	}
	return sitter.NewQuery([]byte(importsQuery), typescriptLang)
}
