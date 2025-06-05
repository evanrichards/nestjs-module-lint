package analysis

// ModuleAnalyzer defines the interface for analyzing NestJS modules
type ModuleAnalyzer interface {
	// AnalyzeFile analyzes a single module file
	AnalyzeFile(filePath string) ([]*ModuleAnalysisResult, error)

	// AnalyzeDirectory recursively analyzes all modules in a directory
	AnalyzeDirectory(dirPath string) ([]*ModuleAnalysisResult, error)
}

// PathResolver defines the interface for resolving import paths
type PathResolver interface {
	ResolveImportPath(baseDir, importPath string) string
}

// IgnoreDetector defines the interface for handling ignore comments
type IgnoreDetector interface {
	ShouldIgnoreFile(source []byte) bool
	ShouldIgnoreImport(moduleName string, source []byte) bool
}

// ReExportDetector defines the interface for detecting re-exports
type ReExportDetector interface {
	GetReExportedModules(imports []string, exports []string) []string
}

// ModuleParser defines the interface for parsing module information
type ModuleParser interface {
	ParseModuleInfo(filePath string) (*ModuleInfo, error)
	GetImportsByModule(filePath string) (map[string][]string, error)
	GetExportsByModule(filePath string) (map[string][]string, error)
	GetProvidersByModule(filePath string) (map[string][]string, error)
}
