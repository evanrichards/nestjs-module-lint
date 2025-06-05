package app

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	resolver "github.com/evanrichards/nestjs-module-lint/internal/resolver"
	sitter "github.com/smacker/go-tree-sitter"
)

// ClassInheritanceInfo represents inheritance information for a class
type ClassInheritanceInfo struct {
	ClassName      string
	BaseClass      string
	HasConstructor bool
}

// AnalyzeClassInheritance analyzes class inheritance in TypeScript source code
func AnalyzeClassInheritance(sourceCode []byte) ([]ClassInheritanceInfo, error) {
	// Use a simpler regex-based approach for now since tree-sitter query is complex
	source := string(sourceCode)
	var inheritanceInfo []ClassInheritanceInfo

	// Look for class declarations with extends
	classPattern := regexp.MustCompile(`export\s+class\s+(\w+)\s+extends\s+(\w+)\s*{([^}]*)`)
	matches := classPattern.FindAllStringSubmatch(source, -1)

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
	nonExportPattern := regexp.MustCompile(`(?:^|\n)\s*class\s+(\w+)\s+extends\s+(\w+)\s*{([^}]*)`)
	nonExportMatches := nonExportPattern.FindAllStringSubmatch(source, -1)

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

// getInheritedDependencies recursively finds dependencies from base classes
func getInheritedDependencies(className string, filePath string, pathResolver *resolver.TsPathResolver, visited map[string]bool) ([]FileImportNode, error) {
	// Prevent infinite recursion
	if visited[filePath] {
		return nil, nil
	}
	visited[filePath] = true

	// Read and analyze the file
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		// If we can't read the file, just return empty dependencies
		return nil, nil
	}

	// Analyze inheritance in this file
	inheritanceInfo, err := AnalyzeClassInheritance(sourceCode)
	if err != nil {
		return nil, err
	}

	// Find inheritance info for our target class
	var targetInheritance *ClassInheritanceInfo
	for _, info := range inheritanceInfo {
		if info.ClassName == className {
			targetInheritance = &info
			break
		}
	}

	// If no inheritance or has explicit constructor, get direct dependencies
	if targetInheritance == nil || targetInheritance.BaseClass == "" || targetInheritance.HasConstructor {
		return getFileImports(filePath, pathResolver)
	}

	// Get dependencies from base class
	baseClassFile := findBaseClassFile(targetInheritance.BaseClass, filePath, pathResolver)
	if baseClassFile == "" {
		// If we can't find the base class file, fall back to direct dependencies
		return getFileImports(filePath, pathResolver)
	}

	// Recursively get dependencies from base class
	baseDependencies, err := getInheritedDependencies(targetInheritance.BaseClass, baseClassFile, pathResolver, visited)
	if err != nil {
		return nil, err
	}

	// Also get any direct dependencies from this file (other imports)
	directDependencies, err := getFileImports(filePath, pathResolver)
	if err != nil {
		return nil, err
	}

	// Combine dependencies, avoiding duplicates
	dependencyMap := make(map[string]FileImportNode)

	for _, dep := range baseDependencies {
		dependencyMap[dep.name] = dep
	}

	for _, dep := range directDependencies {
		// Skip the base class import itself
		if dep.name != targetInheritance.BaseClass {
			dependencyMap[dep.name] = dep
		}
	}

	// Convert back to slice
	var allDependencies []FileImportNode
	for _, dep := range dependencyMap {
		allDependencies = append(allDependencies, dep)
	}

	return allDependencies, nil
}

// findBaseClassFile finds the file containing the base class
func findBaseClassFile(baseClassName string, currentFile string, pathResolver *resolver.TsPathResolver) string {
	// Read the current file to find the import statement for the base class
	sourceCode, err := os.ReadFile(currentFile)
	if err != nil {
		return ""
	}

	// Parse to find imports
	tree, err := sitter.ParseCtx(context.Background(), sourceCode, getTypescriptLanguage())
	if err != nil {
		return ""
	}

	// Get import paths
	importPaths, err := parser.ParseImportPaths(tree, sourceCode)
	if err != nil {
		return ""
	}

	// Find the import path for our base class
	baseClassPath, exists := importPaths[baseClassName]
	if !exists {
		return ""
	}

	// Resolve the full path
	fileDir := filepath.Dir(currentFile)
	return pathResolver.ResolveImportPath(fileDir, baseClassPath)
}

// getFileImports is a helper to get file imports without inheritance analysis
func getFileImports(filePath string, pathResolver *resolver.TsPathResolver) ([]FileImportNode, error) {
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := sitter.ParseCtx(context.Background(), sourceCode, getTypescriptLanguage())
	if err != nil {
		return nil, err
	}

	return getFileImportsFromAST(tree, sourceCode, pathResolver, filePath)
}
