package parser

import (
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func ParseAll(
	sourceCode []byte,
) error {
	// Parse source code
	lang := typescript.GetLanguage()
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil {
		return err
	}
	moduleInputPairs, err := GetImportsByModuleFromFile(n, sourceCode)
	if err != nil {
		return err
	}
	fmt.Printf("%v", moduleInputPairs)
	return nil
}
