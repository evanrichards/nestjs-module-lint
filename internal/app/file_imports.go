package app

import (
	"path/filepath"
	"strings"

	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	"github.com/evanrichards/nestjs-module-lint/internal/resolver"
	sitter "github.com/smacker/go-tree-sitter"
)

// getFileImportsFromAST extracts import information from a parsed AST node
func getFileImportsFromAST(
	n *sitter.Node,
	sourceCode []byte,
	pathResolver *resolver.TsPathResolver,
	filePath string,
) ([]FileImportNode, error) {
	fileImports, err := parser.ParseImportPaths(n, sourceCode)
	if err != nil {
		return nil, err
	}

	var fileImportNodes []FileImportNode
	fileDir := filepath.Dir(filePath)
	for importName, importPath := range fileImports {
		// Skip @nestjs/ imports as they are framework imports
		if strings.HasPrefix(importPath, "@nestjs/") {
			continue
		}
		fullpath := pathResolver.ResolveImportPath(fileDir, importPath)
		fileImportNodes = append(fileImportNodes, FileImportNode{
			path:     importPath,
			name:     importName,
			fullpath: fullpath,
		})
	}
	return fileImportNodes, nil
}
