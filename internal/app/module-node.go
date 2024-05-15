package app

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/loop-payments/nestjs-module-lint/internal/parser"
	pathresolver "github.com/loop-payments/nestjs-module-lint/internal/path-resolver"
	sitter "github.com/smacker/go-tree-sitter"
)

type FileImportNode struct {
	path     string
	name     string
	fullpath string
}

type ModuleImportNode struct {
	path     string
	name     string
	fullpath string
	exports  []string
}

type ProviderControllerNode struct {
	name        string
	path        string
	fullpath    string
	fileImports []FileImportNode
}

type ModuleNode struct {
	name                string
	imports             []ModuleImportNode
	providerControllers []ProviderControllerNode
	pathResolver        *pathresolver.TsPathResolver
}

func findFileImportNodeByName(fileImportNodes []FileImportNode, name string) (*FileImportNode, bool) {
	for _, fileImportNode := range fileImportNodes {
		if fileImportNode.name == name {
			return &fileImportNode, true
		}
	}
	return nil, false
}

func NewModuleNode(name string, importNames, providerControllerNames []string, fileImportNodes []FileImportNode, pathResolver *pathresolver.TsPathResolver) *ModuleNode {
	var filteredImports []ModuleImportNode
	for _, importName := range importNames {
		if fileImportNode, ok := findFileImportNodeByName(fileImportNodes, importName); ok {
			filteredImports = append(filteredImports, ModuleImportNode{fileImportNode.path, importName, fileImportNode.fullpath, nil})
		}
	}

	var filteredProviderControllers []ProviderControllerNode
	for _, providerControllerName := range providerControllerNames {
		if fileImportNode, ok := findFileImportNodeByName(fileImportNodes, providerControllerName); ok {
			filteredProviderControllers = append(filteredProviderControllers, ProviderControllerNode{providerControllerName, fileImportNode.path, fileImportNode.fullpath, nil})
		}
	}

	return &ModuleNode{
		name:                name,
		imports:             filteredImports,
		providerControllers: filteredProviderControllers,
		pathResolver:        pathResolver,
	}
}

func (m *ModuleNode) Check() ([]string, error) {
	var unnecessaryImports []string
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorChan := make(chan error, 1)

	for i := range m.imports {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			exports, err := getExportsForModule(m.imports[i].name, m.imports[i].fullpath)
			if err != nil {
				select {
				case errorChan <- err:
				default:
				}
				return
			}
			mu.Lock()
			if exports == nil {
				unnecessaryImports = append(unnecessaryImports, m.imports[i].name)
			}
			m.imports[i].exports = exports
			mu.Unlock()
		}(i)
	}

	for i := range m.providerControllers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fileImports, err := getFileImportsForFile(m.providerControllers[i].fullpath, m.pathResolver)
			if err != nil {
				select {
				case errorChan <- err:
				default:
				}
				return
			}
			mu.Lock()
			m.providerControllers[i].fileImports = fileImports
			mu.Unlock()
		}(i)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	if err := <-errorChan; err != nil {
		return nil, err
	}

	for _, importNode := range m.imports {
		found := false
		for _, export := range importNode.exports {
			if found {
				break
			}
			for _, providerController := range m.providerControllers {
				if _, ok := findFileImportNodeByName(providerController.fileImports, export); ok {
					found = true
					break
				}
			}
		}
		if !found {
			unnecessaryImports = append(unnecessaryImports, importNode.name)
		}
	}
	return unnecessaryImports, nil
}

func getFileImportsForFile(filePath string, pathResolver *pathresolver.TsPathResolver) ([]FileImportNode, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil {
		return nil, err
	}
	return getFileImports(n, sourceCode, pathResolver, filePath)
}

func getFileImports(
	n *sitter.Node,
	sourceCode []byte,
	pathResolver *pathresolver.TsPathResolver,
	filePath string,
) ([]FileImportNode, error) {
	fileImports, err := parser.GetImportPathsByImportNames(n, sourceCode)
	if err != nil {
		return nil, err
	}
	var fileImportNodes []FileImportNode
	fileDir := filepath.Dir(filePath)
	for importName, importPath := range fileImports {
		fullpath := pathResolver.ResolveImportPath(fileDir, importPath)
		fileImportNodes = append(fileImportNodes, FileImportNode{importPath, importName, fullpath})
	}
	return fileImportNodes, nil
}

func getExportsForModule(moduleName, filePath string) ([]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	n, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
	if err != nil {
		return nil, err
	}
	exportsByModule, err := parser.GetExportsByModuleFromFile(n, sourceCode)
	if err != nil {
		return nil, err
	}
	exports, ok := exportsByModule[moduleName]
	if !ok {
		return nil, nil
	}
	return exports, nil
}
