package detection

import "github.com/evanrichards/nestjs-module-lint/internal/analysis"

// IgnoreDetectorAdapter adapts IgnoreDetector to implement the analysis.IgnoreDetector interface
type IgnoreDetectorAdapter struct {
	detector *IgnoreDetector
}

// NewIgnoreDetectorAdapter creates a new ignore detector adapter
func NewIgnoreDetectorAdapter(detector *IgnoreDetector) analysis.IgnoreDetector {
	return &IgnoreDetectorAdapter{
		detector: detector,
	}
}

// ShouldIgnoreFile implements the analysis.IgnoreDetector interface
func (a *IgnoreDetectorAdapter) ShouldIgnoreFile(source []byte) bool {
	return a.detector.ShouldIgnoreFile(source)
}

// ShouldIgnoreImport implements the analysis.IgnoreDetector interface
func (a *IgnoreDetectorAdapter) ShouldIgnoreImport(moduleName string, source []byte) bool {
	return a.detector.ShouldIgnoreImport(moduleName, source)
}

// ReExportDetectorAdapter adapts ReExportDetector to implement the analysis.ReExportDetector interface
type ReExportDetectorAdapter struct {
	detector *ReExportDetector
}

// NewReExportDetectorAdapter creates a new re-export detector adapter
func NewReExportDetectorAdapter(detector *ReExportDetector) analysis.ReExportDetector {
	return &ReExportDetectorAdapter{
		detector: detector,
	}
}

// GetReExportedModules implements the analysis.ReExportDetector interface
func (a *ReExportDetectorAdapter) GetReExportedModules(imports []string, exports []string) []string {
	return a.detector.GetReExportedModules(imports, exports)
}
