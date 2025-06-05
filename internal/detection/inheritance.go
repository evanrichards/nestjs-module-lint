package detection

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	pathresolver "github.com/evanrichards/nestjs-module-lint/internal/path-resolver"
	sitter "github.com/smacker/go-tree-sitter"
)

// ClassInheritanceInfo represents inheritance information for a class
type ClassInheritanceInfo struct {
	ClassName      string
	BaseClass      string
	HasConstructor bool
}

// InheritanceAnalyzer handles class inheritance analysis
type InheritanceAnalyzer struct {
	classPattern     *regexp.Regexp
	nonExportPattern *regexp.Regexp
	lang             *sitter.Language
}

// NewInheritanceAnalyzer creates a new inheritance analyzer
func NewInheritanceAnalyzer(lang *sitter.Language) *InheritanceAnalyzer {
	return &InheritanceAnalyzer{
		classPattern:     regexp.MustCompile(`export\s+class\s+(\w+)\s+extends\s+(\w+)\s*{([^}]*)`),
		nonExportPattern: regexp.MustCompile(`(?:^|\n)\s*class\s+(\w+)\s+extends\s+(\w+)\s*{([^}]*)`),
		lang:             lang,
	}
}

// AnalyzeClassInheritance analyzes class inheritance in TypeScript source code
func (a *InheritanceAnalyzer) AnalyzeClassInheritance(sourceCode []byte) ([]ClassInheritanceInfo, error) {
	source := string(sourceCode)
	var inheritanceInfo []ClassInheritanceInfo

	// Look for exported class declarations with extends
	matches := a.classPattern.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		if len(match) >= 4 {
			className := match[1]
			baseClass := match[2]
			classBody := match[3]

			hasConstructor := strings.Contains(classBody, "constructor(")

			inheritanceInfo = append(inheritanceInfo, ClassInheritanceInfo{
				ClassName:      className,
				BaseClass:      baseClass,
				HasConstructor: hasConstructor,
			})
		}
	}

	// Also check for non-exported classes
	nonExportMatches := a.nonExportPattern.FindAllStringSubmatch(source, -1)
	for _, match := range nonExportMatches {
		if len(match) >= 4 {
			className := match[1]
			baseClass := match[2]
			classBody := match[3]

			hasConstructor := strings.Contains(classBody, "constructor(")

			inheritanceInfo = append(inheritanceInfo, ClassInheritanceInfo{
				ClassName:      className,
				BaseClass:      baseClass,
				HasConstructor: hasConstructor,
			})
		}
	}

	return inheritanceInfo, nil
}

// GetInheritedDependencies recursively finds dependencies through inheritance chains
func (a *InheritanceAnalyzer) GetInheritedDependencies(
	className string,
	filePath string,
	pathResolver *pathresolver.TsPathResolver,
	visited map[string]bool,
) ([]FileImportNode, error) {
	// Prevent infinite recursion
	if visited[className] {
		return []FileImportNode{}, nil
	}
	visited[className] = true

	// Read the file containing the class
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Get inheritance info for this file
	inheritanceInfos, err := a.AnalyzeClassInheritance(sourceCode)
	if err != nil {
		return nil, err
	}

	// Find inheritance info for our target class
	var targetInheritance *ClassInheritanceInfo
	for _, info := range inheritanceInfos {
		if info.ClassName == className {
			targetInheritance = &info
			break
		}
	}

	// Parse the file to get imports
	n, err := sitter.ParseCtx(context.Background(), sourceCode, a.lang)
	if err != nil {
		return nil, err
	}

	var allDependencies []FileImportNode

	// Get direct imports from this file
	directImports, err := getFileImportsFromNode(n, sourceCode, pathResolver, filePath)
	if err != nil {
		return nil, err
	}
	allDependencies = append(allDependencies, directImports...)

	// If this class extends another class and doesn't have its own constructor,
	// it inherits the base class's dependencies
	if targetInheritance != nil && !targetInheritance.HasConstructor {
		// Find the base class file
		baseClassFile := findBaseClassFile(targetInheritance.BaseClass, directImports, pathResolver)
		if baseClassFile != "" {
			// Recursively get dependencies from the base class
			baseDependencies, err := a.GetInheritedDependencies(
				targetInheritance.BaseClass,
				baseClassFile,
				pathResolver,
				visited,
			)
			if err != nil {
				return nil, err
			}
			allDependencies = append(allDependencies, baseDependencies...)
		}
	}

	return allDependencies, nil
}

// FileImportNode represents a file import
type FileImportNode struct {
	Path     string
	Name     string
	FullPath string
}

// getFileImportsFromNode extracts file imports from AST node
func getFileImportsFromNode(
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
		// Skip @nestjs/ imports
		if strings.HasPrefix(importPath, "@nestjs/") {
			continue
		}
		fullpath := pathResolver.ResolveImportPath(fileDir, importPath)
		fileImportNodes = append(fileImportNodes, FileImportNode{
			Path:     importPath,
			Name:     importName,
			FullPath: fullpath,
		})
	}

	return fileImportNodes, nil
}

// findBaseClassFile finds the file containing the base class
func findBaseClassFile(baseClassName string, imports []FileImportNode, pathResolver *pathresolver.TsPathResolver) string {
	for _, imp := range imports {
		if imp.Name == baseClassName {
			return imp.FullPath
		}
	}
	return ""
}
