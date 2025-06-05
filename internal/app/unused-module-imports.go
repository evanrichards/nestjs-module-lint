package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
	"github.com/evanrichards/nestjs-module-lint/internal/detection"
	"github.com/evanrichards/nestjs-module-lint/internal/filesystem"
	"github.com/evanrichards/nestjs-module-lint/internal/parser"
	"github.com/evanrichards/nestjs-module-lint/internal/resolver"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	mpb "github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

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
// This is the main entry point that bridges old and new architectures
func AnalyzePath(path string) ([]*ModuleReport, error) {
	return analyzePathInternal(path, false)
}

// AnalyzePathWithNewArchitecture uses the new analysis architecture (when fully implemented)
func AnalyzePathWithNewArchitecture(path string) ([]*ModuleReport, error) {
	return analyzePathInternal(path, true)
}

// analyzePathInternal contains the actual implementation with architecture selection
func analyzePathInternal(path string, useNewArchitecture bool) ([]*ModuleReport, error) {
	if useNewArchitecture {
		return analyzeWithNewArchitecture(path)
	}
	return analyzeWithLegacyArchitecture(path)
}

// analyzeWithNewArchitecture uses the new analysis package
func analyzeWithNewArchitecture(path string) ([]*ModuleReport, error) {
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

// analyzeWithLegacyArchitecture uses the existing implementation
func analyzeWithLegacyArchitecture(path string) ([]*ModuleReport, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist")
		}
		return nil, fmt.Errorf("cannot access path: %w", err)
	}

	var files []string
	if info.IsDir() {
		files, err = filesystem.FindTypeScriptFiles(path)
		if err != nil {
			return nil, fmt.Errorf("failed to find TypeScript files: %w", err)
		}
	} else {
		// Validate file extension for single files
		lowerPath := strings.ToLower(path)
		if !strings.HasSuffix(lowerPath, ".ts") && !strings.HasSuffix(lowerPath, ".tsx") {
			return nil, fmt.Errorf("file must have .ts or .tsx extension")
		}
		files = []string{path}
	}
	p := mpb.New(mpb.WithWidth(64))

	bar := p.New(int64(len(files)),
		// BarFillerBuilder with custom style
		mpb.BarStyle(),
		mpb.PrependDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done"),
		),
		mpb.AppendDecorators(decor.Percentage()),
	)
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		*ModuleReport
		error
	})
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer func() {
				wg.Done()
				bar.Increment()
			}()

			moduleReports, err := AnalyzeModuleFile(file)
			if err != nil {
				resultChan <- struct {
					*ModuleReport
					error
				}{nil, fmt.Errorf("failed to run app for %s: %w", file, err)}
				return
			}

			for _, report := range moduleReports {
				resultChan <- struct {
					*ModuleReport
					error
				}{report, nil}
			}
		}(file)
	}

	// Close the result channel once all goroutines have completed
	go func() {
		wg.Wait()
		p.Wait()
		close(resultChan)
	}()

	var results []*ModuleReport
	for result := range resultChan {
		if result.error != nil {
			return nil, result.error
		}
		if result.ModuleReport != nil && len(result.ModuleReport.UnnecessaryImports) > 0 {
			results = append(results, result.ModuleReport)
		}
	}

	// sort results by module name
	sort.Slice(results, func(i, j int) bool {
		return results[i].ModuleName < results[j].ModuleName
	})

	return results, nil
}

// AnalyzeModuleFile analyzes a single TypeScript module file for unused imports
func AnalyzeModuleFile(filePath string) ([]*ModuleReport, error) {
	qualifiedPathToModule, cwd, err := resolveFilePath(filePath)
	if err != nil {
		return nil, err
	}

	pathResolver, err := resolver.NewTsPathResolverFromPath(cwd)
	if err != nil {
		return nil, err
	}
	sourceCode, err := os.ReadFile(qualifiedPathToModule)
	if err != nil {
		return nil, errors.Join(errors.New("could not read the input file, does it exist?"), err)
	}

	// Parse ignore comments
	ignoreDetector := detection.NewIgnoreDetector()
	ignoreInfo := ignoreDetector.ParseIgnoreComments(sourceCode)

	// If the entire file is ignored, return empty results
	if ignoreInfo.FileIgnored {
		return []*ModuleReport{}, nil
	}

	// Initialize re-export detector
	reExportDetector := detection.NewReExportDetector()

	n, err := sitter.ParseCtx(context.Background(), sourceCode, getTypescriptLanguage())
	if err != nil {
		return nil, errors.Join(errors.New("could not parse the input file, is it valid typescript?"), err)
	}
	importsByModule, err := parser.ParseModuleImports(n, sourceCode)
	if err != nil {
		return nil, err
	}
	fileImports, err := getFileImportsFromAST(n, sourceCode, pathResolver, qualifiedPathToModule)
	if err != nil {
		return nil, err
	}
	providerControllersByModule, err := parser.ParseModuleProviders(n, sourceCode)
	if err != nil {
		return nil, err
	}
	exportsByModule, err := parser.ParseModuleExports(n, sourceCode)
	if err != nil {
		return nil, err
	}

	moduleReports := make([]*ModuleReport, 0)
	for module, imports := range importsByModule {
		providerControllers, ok := providerControllersByModule[module]

		// Get exports for this module to check for re-export patterns
		moduleExports, hasExports := exportsByModule[module]

		if !ok {
			// Convert absolute path to relative path from project root
			relativePath := getRelativePath(qualifiedPathToModule, cwd)

			// Filter out ignored imports
			filteredImports := ignoreDetector.GetNonIgnoredImports(imports, sourceCode)

			// Filter out re-exported imports
			if hasExports {
				filteredImports = reExportDetector.GetNonReExportedImports(filteredImports, moduleExports)
			}

			// Only create a report if there are still unused imports after filtering
			if len(filteredImports) > 0 {
				moduleReports = append(moduleReports, &ModuleReport{
					ModuleName:         module,
					Path:               relativePath,
					UnnecessaryImports: filteredImports,
				})
			}
			continue
		}

		var moduleExportsForModule []string
		if hasExports {
			moduleExportsForModule = moduleExports
		}

		moduleReport, err := analyzeModule(module, imports, providerControllers, fileImports, pathResolver, qualifiedPathToModule, ignoreInfo, moduleExportsForModule)
		if err != nil {
			return nil, err
		}
		if moduleReport != nil && len(moduleReport.UnnecessaryImports) > 0 {
			moduleReports = append(moduleReports, moduleReport)
		}
	}
	return moduleReports, nil
}

type ModuleReport struct {
	ModuleName         string   `json:"module_name"`
	Path               string   `json:"path"`
	UnnecessaryImports []string `json:"unnecessary_imports"`
}

func analyzeModule(
	moduleName string,
	importNames []string,
	providerControllers []string,
	fileImports []FileImportNode,
	pathResolver *resolver.TsPathResolver,
	qualifiedPathToModule string,
	ignoreInfo *detection.IgnoreInfo,
	moduleExports []string,
) (*ModuleReport, error) {
	moduleNode := NewModuleNode(moduleName, importNames, providerControllers, fileImports, pathResolver)
	unnecessaryInputs, err := moduleNode.Check()
	if err != nil {
		return nil, err
	}

	// Filter out ignored imports using the ignoreInfo that was passed in
	// Note: We need the source code to properly filter, but for now we'll use the passed ignoreInfo
	var filteredImports []string
	for _, imp := range unnecessaryInputs {
		if !ignoreInfo.ShouldIgnoreModule(imp) {
			filteredImports = append(filteredImports, imp)
		}
	}

	// Filter out re-exported imports
	if len(moduleExports) > 0 {
		reExportDetector := detection.NewReExportDetector()
		filteredImports = reExportDetector.GetNonReExportedImports(filteredImports, moduleExports)
	}

	// Convert absolute path to relative path from project root
	cwd, err := getWorkingDirectory()
	if err != nil {
		// If we can't get working directory, just use the qualified path
		return &ModuleReport{
			ModuleName:         moduleName,
			Path:               qualifiedPathToModule,
			UnnecessaryImports: filteredImports,
		}, nil
	}
	relativePath := getRelativePath(qualifiedPathToModule, cwd)

	return &ModuleReport{
		ModuleName:         moduleName,
		Path:               relativePath,
		UnnecessaryImports: filteredImports,
	}, nil
}

func PrettyPrintModuleReport(report *ModuleReport) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Module: %s\nPath: %s\nUnnecessary Imports:\n", report.ModuleName, report.Path))
	for _, imp := range report.UnnecessaryImports {
		builder.WriteString(fmt.Sprintf("\t%s\n", imp))
	}
	return builder.String()
}
