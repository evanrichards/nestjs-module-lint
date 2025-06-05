package analysis

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/evanrichards/nestjs-module-lint/internal/filesystem"
)

// moduleImportData holds information about an imported module
type moduleImportData struct {
	name    string
	path    string
	exports []string
}

// providerData holds information about a provider/controller
type providerData struct {
	name        string
	path        string
	fileImports []string
}

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
	if len(providers) == 0 {
		// If there are no providers/controllers, all imports are potentially unused
		// However, this is a conservative check - modules might still be used in other ways
		return imports
	}

	// Build the dependency map for concurrent analysis
	importData := make([]moduleImportData, len(imports))
	for i, importName := range imports {
		importData[i] = moduleImportData{
			name: importName,
			path: a.pathResolver.ResolveImportPath(filepath.Dir(filePath), importName),
		}
	}

	providerList := make([]providerData, len(providers))
	for i, providerName := range providers {
		providerList[i] = providerData{
			name: providerName,
			path: a.pathResolver.ResolveImportPath(filepath.Dir(filePath), providerName),
		}
	}

	// Perform concurrent analysis
	return a.analyzeImportUsage(importData, providerList)
}

// analyzeImportUsage performs the actual dependency analysis using concurrent processing
func (a *Analyzer) analyzeImportUsage(imports []moduleImportData, providers []providerData) []string {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorChan := make(chan error, 1)

	// Concurrently get exports for each imported module
	for i := range imports {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			exports, err := a.getModuleExports(imports[i].name, imports[i].path)
			if err != nil {
				select {
				case errorChan <- err:
				default:
				}
				return
			}
			mu.Lock()
			imports[i].exports = exports
			mu.Unlock()
		}(i)
	}

	// Concurrently get file imports for each provider/controller
	for i := range providers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fileImports, err := a.getProviderFileImports(providers[i].path)
			if err != nil {
				select {
				case errorChan <- err:
				default:
				}
				return
			}
			mu.Lock()
			providers[i].fileImports = fileImports
			mu.Unlock()
		}(i)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Check for errors
	if err := <-errorChan; err != nil {
		// On error, conservatively assume all imports are used
		return []string{}
	}

	// Build a set of all imports used by providers
	usedImports := make(map[string]bool)
	for _, provider := range providers {
		for _, fileImport := range provider.fileImports {
			usedImports[fileImport] = true
		}
	}

	// Check which imported module exports are actually used
	var unusedImports []string
	for _, importModule := range imports {
		found := false
		for _, export := range importModule.exports {
			if usedImports[export] {
				found = true
				break
			}
		}
		if !found {
			unusedImports = append(unusedImports, importModule.name)
		}
	}

	return unusedImports
}

// getModuleExports gets the exports from a module file
func (a *Analyzer) getModuleExports(moduleName, filePath string) ([]string, error) {
	result, err := a.parser.GetExportsByModule(filePath)
	if err != nil {
		return nil, err
	}

	// Return exports for this specific module
	if exports, exists := result[moduleName]; exists {
		return exports, nil
	}

	// If no exports found for this module, it might not export anything
	return []string{}, nil
}

// getProviderFileImports gets the file imports for a provider/controller file
func (a *Analyzer) getProviderFileImports(filePath string) ([]string, error) {
	importPaths, err := a.parser.GetImportPaths(filePath)
	if err != nil {
		return nil, err
	}

	// Extract just the import names (not the full paths)
	var importNames []string
	for importName := range importPaths {
		// Skip @nestjs/ imports as they're not relevant for module dependency analysis
		if !strings.HasPrefix(importPaths[importName], "@nestjs/") {
			importNames = append(importNames, importName)
		}
	}

	return importNames, nil
}
