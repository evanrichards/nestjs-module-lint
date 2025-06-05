package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
	"github.com/evanrichards/nestjs-module-lint/internal/detection"
	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	"github.com/evanrichards/nestjs-module-lint/internal/resolver"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// FileImportNode represents a file import for compatibility with inheritance analysis
type FileImportNode struct {
	path     string
	name     string
	fullpath string
}

// getTypescriptLanguage returns the TypeScript language instance
func getTypescriptLanguage() *sitter.Language {
	return typescript.GetLanguage()
}

// getWorkingDirectory returns the current working directory
func getWorkingDirectory() (string, error) {
	return os.Getwd()
}

// resolveFilePath converts relative paths to absolute paths based on working directory
func resolveFilePath(filePath string) (string, string, error) {
	cwd, err := getWorkingDirectory()
	if err != nil {
		return "", "", err
	}

	var qualifiedPath string
	if filepath.IsAbs(filePath) {
		qualifiedPath = filePath
	} else {
		qualifiedPath = filepath.Join(cwd, filePath)
	}

	return qualifiedPath, cwd, nil
}

// getRelativePath returns a relative path from the working directory, falling back to absolute path if conversion fails
func getRelativePath(absolutePath, workingDir string) string {
	relativePath, err := filepath.Rel(workingDir, absolutePath)
	if err != nil {
		// If we can't get relative path, fall back to the original path
		return absolutePath
	}
	return relativePath
}

// AnalyzePath analyzes a file or directory for unused module imports
// This is the main entry point using the new analysis architecture
func AnalyzePath(path string) ([]*ModuleReport, error) {
	// Get current working directory
	cwd, err := getWorkingDirectory()
	if err != nil {
		return nil, err
	}

	// Create path resolver
	tsPathResolver, err := resolver.NewTsPathResolverFromPath(cwd)
	if err != nil {
		return nil, err
	}
	pathResolverAdapter := resolver.NewPathResolverAdapter(tsPathResolver)

	// Create parser adapter
	parserAdapter := parser.NewParserAdapter(getTypescriptLanguage())

	// Create detection adapters
	ignoreDetector := detection.NewIgnoreDetector()
	ignoreAdapter := detection.NewIgnoreDetectorAdapter(ignoreDetector)

	reExportDetector := detection.NewReExportDetector()
	reExportAdapter := detection.NewReExportDetectorAdapter(reExportDetector)

	// Create analysis options
	options := analysis.AnalysisOptions{
		WorkingDirectory: cwd,
		EnableIgnores:    true,
		EnableReExports:  true,
	}

	// Create analyzer
	analyzer := analysis.NewAnalyzer(
		parserAdapter,
		pathResolverAdapter,
		ignoreAdapter,
		reExportAdapter,
		options,
	)

	// Determine if we're analyzing a file or directory
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var results []*analysis.ModuleAnalysisResult
	if info.IsDir() {
		results, err = analyzer.AnalyzeDirectory(path)
	} else {
		results, err = analyzer.AnalyzeFile(path)
	}

	if err != nil {
		return nil, err
	}

	// Convert analysis results to ModuleReport for backward compatibility
	var reports []*ModuleReport
	for _, result := range results {
		if len(result.UnusedImports) > 0 {
			reports = append(reports, &ModuleReport{
				ModuleName:         result.ModuleName,
				Path:               result.FilePath,
				UnnecessaryImports: result.UnusedImports,
			})
		}
	}

	return reports, nil
}

type ModuleReport struct {
	ModuleName         string   `json:"module_name"`
	Path               string   `json:"path"`
	UnnecessaryImports []string `json:"unnecessary_imports"`
}

func PrettyPrintModuleReport(report *ModuleReport) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Module: %s\nPath: %s\nUnnecessary Imports:\n", report.ModuleName, report.Path))
	for _, imp := range report.UnnecessaryImports {
		builder.WriteString(fmt.Sprintf("\t%s\n", imp))
	}
	return builder.String()
}
