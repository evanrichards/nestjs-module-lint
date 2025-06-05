package detection

// ReExportDetector implements re-export pattern detection
type ReExportDetector struct{}

// NewReExportDetector creates a new re-export detector
func NewReExportDetector() *ReExportDetector {
	return &ReExportDetector{}
}

// GetReExportedModules returns modules from imports list that are also exported (re-export pattern)
func (d *ReExportDetector) GetReExportedModules(imports []string, exports []string) []string {
	if len(exports) == 0 {
		return []string{} // No exports, so no re-exports
	}

	// Create a set of exported modules for efficient lookup
	exportedModules := make(map[string]bool)
	for _, export := range exports {
		exportedModules[export] = true
	}

	// Find imports that are also exported
	var reExported []string
	for _, imp := range imports {
		if exportedModules[imp] {
			reExported = append(reExported, imp)
		}
	}

	return reExported
}

// GetNonReExportedImports returns modules from imports list that are NOT re-exported
func (d *ReExportDetector) GetNonReExportedImports(imports []string, exports []string) []string {
	if len(exports) == 0 {
		return imports // No exports, so no re-exports to filter
	}

	// Create a set of exported modules for efficient lookup
	exportedModules := make(map[string]bool)
	for _, export := range exports {
		exportedModules[export] = true
	}

	// Filter out imports that are also exported
	var filtered []string
	for _, imp := range imports {
		if !exportedModules[imp] {
			filtered = append(filtered, imp)
		}
	}

	return filtered
}

// IsReExported checks if a specific module is re-exported
func (d *ReExportDetector) IsReExported(moduleName string, exports []string) bool {
	for _, export := range exports {
		if export == moduleName {
			return true
		}
	}
	return false
}
