package parser

import (
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"path/filepath"
)

var typescriptLang = typescript.GetLanguage()

func LoadModuleImportQuery() (*sitter.Query, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	importsQuery, err := os.ReadFile(filepath.Join(cwd, "internal", "parser", "module-imports.query"))
	if err != nil {
		return nil, err
	}
	return sitter.NewQuery([]byte(importsQuery), typescriptLang)
}
