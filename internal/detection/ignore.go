package detection

import (
	"regexp"
	"strings"
)

// IgnoreDetector implements ignore comment detection
type IgnoreDetector struct {
	fileIgnorePattern *regexp.Regexp
	lineIgnorePattern *regexp.Regexp
	moduleNamePattern *regexp.Regexp
}

// IgnoreInfo contains information about ignore comments in a source file
type IgnoreInfo struct {
	FileIgnored    bool
	IgnoredLines   map[int]bool    // line numbers that have disable-line comments
	IgnoredModules map[string]bool // module names that should be ignored
}

// NewIgnoreDetector creates a new ignore comment detector
func NewIgnoreDetector() *IgnoreDetector {
	return &IgnoreDetector{
		fileIgnorePattern: regexp.MustCompile(`//\s*nestjs-module-lint-disable-file`),
		lineIgnorePattern: regexp.MustCompile(`//\s*nestjs-module-lint-disable-line`),
		moduleNamePattern: regexp.MustCompile(`(\w+),?\s*//\s*nestjs-module-lint-disable-line`),
	}
}

// ShouldIgnoreFile checks if the entire file should be ignored
func (d *IgnoreDetector) ShouldIgnoreFile(source []byte) bool {
	return d.fileIgnorePattern.Match(source)
}

// ShouldIgnoreImport checks if a specific import should be ignored
func (d *IgnoreDetector) ShouldIgnoreImport(moduleName string, source []byte) bool {
	info := d.ParseIgnoreComments(source)
	return info.ShouldIgnoreModule(moduleName)
}

// ParseIgnoreComments analyzes source code for ignore comments
func (d *IgnoreDetector) ParseIgnoreComments(sourceCode []byte) *IgnoreInfo {
	source := string(sourceCode)
	lines := strings.Split(source, "\n")

	info := &IgnoreInfo{
		FileIgnored:    false,
		IgnoredLines:   make(map[int]bool),
		IgnoredModules: make(map[string]bool),
	}

	// Check for file-level ignore
	if d.fileIgnorePattern.MatchString(source) {
		info.FileIgnored = true
		return info // If file is ignored, no need to check line-level ignores
	}

	// Check for line-level ignores
	for i, line := range lines {
		lineNum := i + 1 // Line numbers are 1-based

		if d.lineIgnorePattern.MatchString(line) {
			info.IgnoredLines[lineNum] = true

			// Extract module name from the line if possible
			moduleMatches := d.moduleNamePattern.FindStringSubmatch(line)
			if len(moduleMatches) > 1 {
				moduleName := strings.TrimSpace(moduleMatches[1])
				info.IgnoredModules[moduleName] = true
			}
		}
	}

	return info
}

// ShouldIgnoreModule determines if a module should be ignored based on ignore info
func (info *IgnoreInfo) ShouldIgnoreModule(moduleName string) bool {
	if info.FileIgnored {
		return true
	}

	return info.IgnoredModules[moduleName]
}

// GetIgnoredImports returns a list of ignored imports from the given list
func (d *IgnoreDetector) GetIgnoredImports(imports []string, source []byte) []string {
	info := d.ParseIgnoreComments(source)
	if info.FileIgnored {
		return imports // All imports ignored if file is ignored
	}

	var ignored []string
	for _, imp := range imports {
		if info.ShouldIgnoreModule(imp) {
			ignored = append(ignored, imp)
		}
	}

	return ignored
}

// GetNonIgnoredImports returns a list of non-ignored imports from the given list
func (d *IgnoreDetector) GetNonIgnoredImports(imports []string, source []byte) []string {
	info := d.ParseIgnoreComments(source)
	if info.FileIgnored {
		return []string{} // All imports ignored if file is ignored
	}

	var filtered []string
	for _, imp := range imports {
		if !info.ShouldIgnoreModule(imp) {
			filtered = append(filtered, imp)
		}
	}

	return filtered
}
