package parser

import (
	"context"
	"fmt"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

func ParseAll(
	sourceCode []byte,
) error {
	// Parse source code using the singleton language instance
	n, err := sitter.ParseCtx(context.Background(), sourceCode, typescriptLang)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var moduleInputPairs map[string][]string
	var moduleExportPairs map[string][]string
	var moduleProviderPairs map[string][]string
	var importPathsByName map[string]string
	var err1, err2, err3, err4 error
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		moduleExportPairs, err3 = GetExportsByModuleFromFile(n, sourceCode)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		moduleProviderPairs, err4 = GetProviderControllersByModuleFromFile(n, sourceCode)
	}()

	wg.Wait()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	if err4 != nil {
		return err4
	}

	fmt.Println("Input pairs:")
	fmt.Printf("%v\n", moduleInputPairs)
	fmt.Println("Export pairs:")
	fmt.Printf("%v\n", moduleExportPairs)
	fmt.Println("Provider pairs:")
	fmt.Printf("%v\n", moduleProviderPairs)
	fmt.Println("Import paths:")
	fmt.Printf("%v\n", importPathsByName)
	return nil
}
