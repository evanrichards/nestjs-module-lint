package app

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	resolver "github.com/evanrichards/nestjs-module-lint/internal/resolver"
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
	pathResolver        *resolver.TsPathResolver
}

func findFileImportNodeByName(fileImportNodes []FileImportNode, name string) (*FileImportNode, bool) {
	for _, fileImportNode := range fileImportNodes {
		if fileImportNode.name == name {
			return &fileImportNode, true
		}
	}
	return nil, false
}

func NewModuleNode(name string, importNames, providerControllerNames []string, fileImportNodes []FileImportNode, pathResolver *resolver.TsPathResolver) *ModuleNode {
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

	// Create hash maps for faster lookups
	allProviderImports := make(map[string]bool)
	for _, providerController := range m.providerControllers {
		for _, fileImport := range providerController.fileImports {
			allProviderImports[fileImport.name] = true
		}
	}

	// Check each import for usage
	for _, importNode := range m.imports {
		found := false
		for _, export := range importNode.exports {
			if allProviderImports[export] {
				found = true
				break
			}
		}
		if !found {
			unnecessaryImports = append(unnecessaryImports, importNode.name)
		}
	}
	return unnecessaryImports, nil
}

func getFileImportsForFile(filePath string, pathResolver *resolver.TsPathResolver) ([]FileImportNode, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	n, err := sitter.ParseCtx(context.Background(), sourceCode, getTypescriptLanguage())
	if err != nil {
		return nil, err
	}

	// First, try to find the class name in this file
	className := extractClassNameFromFile(sourceCode)
	if className != "" {
		// Use inheritance-aware dependency analysis
		visited := make(map[string]bool)
		return getInheritedDependencies(className, filePath, pathResolver, visited)
	}

	// Fall back to normal import analysis
	return getFileImportsFromAST(n, sourceCode, pathResolver, filePath)
}

// extractClassNameFromFile finds the first exported class name in a file
func extractClassNameFromFile(sourceCode []byte) string {
	// Look for @Injectable() decorated classes
	lines := strings.Split(string(sourceCode), "\n")
	inInjectableClass := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for @Injectable() decorator
		if strings.Contains(trimmed, "@Injectable()") {
			inInjectableClass = true
			continue
		}

		// If we found @Injectable, look for the next export class declaration
		if inInjectableClass && strings.HasPrefix(trimmed, "export class ") {
			// Extract class name
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				className := parts[2]
				// Remove any '{' or 'extends' parts
				if idx := strings.Index(className, " "); idx != -1 {
					className = className[:idx]
				}
				if idx := strings.Index(className, "{"); idx != -1 {
					className = className[:idx]
				}
				return className
			}
		}

		// Reset if we hit another decorator or class without finding what we need
		if strings.HasPrefix(trimmed, "@") || strings.HasPrefix(trimmed, "export class ") {
			inInjectableClass = false
		}
	}

	return ""
}

func getExportsForModule(moduleName, filePath string) ([]string, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	n, err := sitter.ParseCtx(context.Background(), sourceCode, getTypescriptLanguage())
	if err != nil {
		return nil, err
	}
	exportsByModule, err := parser.ParseModuleExports(n, sourceCode)
	if err != nil {
		return nil, err
	}
	exports, ok := exportsByModule[moduleName]
	if !ok {
		return nil, nil
	}
	return exports, nil
}
