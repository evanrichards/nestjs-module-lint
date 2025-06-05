package analysis

// ModuleAnalysisResult represents the result of analyzing a single module
type ModuleAnalysisResult struct {
	ModuleName        string   `json:"module_name"`
	FilePath          string   `json:"file_path"`
	UnusedImports     []string `json:"unused_imports"`
	IgnoredImports    []string `json:"ignored_imports,omitempty"`
	ReExportedImports []string `json:"reexported_imports,omitempty"`
}

// AnalysisOptions contains configuration for the analysis
type AnalysisOptions struct {
	WorkingDirectory string
	EnableIgnores    bool
	EnableReExports  bool
}

// ModuleInfo contains basic information about a module
type ModuleInfo struct {
	Name      string
	FilePath  string
	Imports   []string
	Exports   []string
	Providers []string
}
