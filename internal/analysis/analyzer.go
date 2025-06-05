package analysis

import (
	"path/filepath"
	"sort"
	"sync"

	"github.com/evanrichards/nestjs-module-lint/internal/filesystem"
)

// Analyzer implements the ModuleAnalyzer interface
type Analyzer struct {
	parser           ModuleParser
	pathResolver     PathResolver
	ignoreDetector   IgnoreDetector
	reExportDetector ReExportDetector
	options          AnalysisOptions
}

// NewAnalyzer creates a new module analyzer with the given dependencies
func NewAnalyzer(
	parser ModuleParser,
	pathResolver PathResolver,
	ignoreDetector IgnoreDetector,
	reExportDetector ReExportDetector,
	options AnalysisOptions,
) *Analyzer {
	return &Analyzer{
		parser:           parser,
		pathResolver:     pathResolver,
		ignoreDetector:   ignoreDetector,
		reExportDetector: reExportDetector,
		options:          options,
	}
}

// AnalyzeFile analyzes a single TypeScript module file
func (a *Analyzer) AnalyzeFile(filePath string) ([]*ModuleAnalysisResult, error) {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Parse module information
	_, err = a.parser.ParseModuleInfo(absPath)
	if err != nil {
		return nil, err
	}

	// Check if file should be ignored
	if a.options.EnableIgnores {
		sourceCode, err := filesystem.ReadFile(absPath)
		if err != nil {
			return nil, err
		}
		if a.ignoreDetector.ShouldIgnoreFile(sourceCode) {
			return []*ModuleAnalysisResult{}, nil
		}
	}

	// Get detailed module data
	importsByModule, err := a.parser.GetImportsByModule(absPath)
	if err != nil {
		return nil, err
	}

	exportsByModule, err := a.parser.GetExportsByModule(absPath)
	if err != nil {
		return nil, err
	}

	providersByModule, err := a.parser.GetProvidersByModule(absPath)
	if err != nil {
		return nil, err
	}

	// Convert to relative path for output
	relativePath, err := filepath.Rel(a.options.WorkingDirectory, absPath)
	if err != nil {
		relativePath = absPath
	}

	var results []*ModuleAnalysisResult
	for moduleName, imports := range importsByModule {
		result := a.analyzeModuleImports(
			moduleName,
			imports,
			exportsByModule[moduleName],
			providersByModule[moduleName],
			relativePath,
			absPath,
		)

		if result != nil && len(result.UnusedImports) > 0 {
			results = append(results, result)
		}
	}

	return results, nil
}

// AnalyzeDirectory recursively analyzes all TypeScript files in a directory
func (a *Analyzer) AnalyzeDirectory(dirPath string) ([]*ModuleAnalysisResult, error) {
	files, err := filesystem.FindTypeScriptFiles(dirPath)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, filesystem.ErrNoTypeScriptFiles
	}

	var allResults []*ModuleAnalysisResult
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			results, err := a.AnalyzeFile(filePath)
			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	if err := <-errChan; err != nil {
		return nil, err
	}

	// Sort results by module name for consistent output
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].ModuleName < allResults[j].ModuleName
	})

	return allResults, nil
}

// analyzeModuleImports analyzes imports for a specific module
func (a *Analyzer) analyzeModuleImports(
	moduleName string,
	imports []string,
	exports []string,
	providers []string,
	relativePath string,
	absolutePath string,
) *ModuleAnalysisResult {
	result := &ModuleAnalysisResult{
		ModuleName:        moduleName,
		FilePath:          relativePath,
		UnusedImports:     make([]string, 0),
		IgnoredImports:    make([]string, 0),
		ReExportedImports: make([]string, 0),
	}

	// Filter ignored imports if enabled
	var filteredImports []string
	if a.options.EnableIgnores {
		sourceCode, err := filesystem.ReadFile(absolutePath)
		if err == nil {
			for _, imp := range imports {
				if a.ignoreDetector.ShouldIgnoreImport(imp, sourceCode) {
					result.IgnoredImports = append(result.IgnoredImports, imp)
				} else {
					filteredImports = append(filteredImports, imp)
				}
			}
		} else {
			filteredImports = imports
		}
	} else {
		filteredImports = imports
	}

	// Filter re-exported imports if enabled
	if a.options.EnableReExports && len(exports) > 0 {
		reExported := a.reExportDetector.GetReExportedModules(filteredImports, exports)
		result.ReExportedImports = reExported

		// Remove re-exported modules from filtered imports
		reExportedSet := make(map[string]bool)
		for _, reExp := range reExported {
			reExportedSet[reExp] = true
		}

		var nonReExported []string
		for _, imp := range filteredImports {
			if !reExportedSet[imp] {
				nonReExported = append(nonReExported, imp)
			}
		}
		filteredImports = nonReExported
	}

	// Analyze actual usage of imports
	result.UnusedImports = a.findUnusedImports(filteredImports, providers, absolutePath)

	return result
}

// findUnusedImports determines which imports are actually unused by analyzing provider dependencies
func (a *Analyzer) findUnusedImports(imports []string, providers []string, filePath string) []string {
	// This is a simplified implementation for Phase 5c
	// The full implementation would need to:
	// 1. Get exports for each imported module (like the current ModuleNode.Check())
	// 2. Get file imports for each provider/controller file
	// 3. Check if the imported modules' exports are actually used

	if len(providers) == 0 {
		// If there are no providers/controllers, we can't determine usage
		// For safety, assume all imports are used (conservative approach)
		return []string{}
	}

	// For this phase, we use a simplified heuristic:
	// If there are providers, assume imports might be used
	// The actual dependency analysis will be implemented in a later phase
	return []string{}
}
