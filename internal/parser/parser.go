package parser

import (
	"context"
	"fmt"
	"sync"

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

	var wg sync.WaitGroup
	var moduleInputPairs map[string][]string
	var importPathsByName map[string]string
	var err1, err2 error
	wg.Add(1)
	go func() {
		defer wg.Done()
		moduleInputPairs, err1 = GetImportsByModuleFromFile(n, sourceCode)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		importPathsByName, err2 = GetImportPathsByImportNames(n, sourceCode)
	}()
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}

	fmt.Printf("%v\n", moduleInputPairs)
	fmt.Printf("%v\n", importPathsByName)
	return nil
}
