package app

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/loop-payments/nestjs-module-lint/internal/parser"
	pathresolver "github.com/loop-payments/nestjs-module-lint/internal/path-resolver"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var cwd string
var lang *sitter.Language

func init() {
	_cwd, err := os.Getwd()
	if err != nil {
		panic("Could not get current file path")
	}
	cwd = _cwd
	lang = typescript.GetLanguage()
}

func Run(
	pathToModule string,
) error {
	qualifiedPathToModule := filepath.Join(cwd, pathToModule)

	pathResolver, err := pathresolver.NewTsPathResolverFromPath(cwd)
	if err != nil {
		return err
	}
	sourceCode, err := os.ReadFile(qualifiedPathToModule)
	if err != nil {
		return errors.Join(errors.New("could not read the input file, does it exist?"), err)
	}
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil {
		return errors.Join(errors.New("could not parse the input file, is it valid typescript?"), err)
	}
	importsByModule, err := parser.GetImportsByModuleFromFile(n, sourceCode)
	if err != nil {
		return err
	}
	fileImports, err := getFileImports(n, sourceCode, pathResolver, qualifiedPathToModule)
	if err != nil {
		return err
	}
	providerControllersByModule, err := parser.GetProviderControllersByModuleFromFile(n, sourceCode)
	if err != nil {
		return err
	}

	for module, imports := range importsByModule {
		providerControllers, ok := providerControllersByModule[module]
		if !ok {
			log.Printf("No provider controllers found for module %s, all imports are unnecessary\n", module)
			continue
		}

		err := runForModule(module, imports, providerControllers, fileImports, pathResolver)
		if err != nil {
			return err
		}
	}
	return nil
}

func runForModule(
	moduleName string,
	importNames []string,
	providerControllers []string,
	fileImports []FileImportNode,
	pathResolver *pathresolver.TsPathResolver,
) error {
	moduleNode := NewModuleNode(moduleName, importNames, providerControllers, fileImports, pathResolver)
	unecessaryInputs, err := moduleNode.Check()
	if err != nil {
		return err
	}
	if len(unecessaryInputs) > 0 {
		log.Printf("%d of %d imports are unnecessary for module %s\n", len(unecessaryInputs), len(importNames), moduleName)
		for _, unecessaryInput := range unecessaryInputs {
			log.Printf(" - %s\n", unecessaryInput)
		}
	}
	return nil
}
