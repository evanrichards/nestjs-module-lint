package app

import (
	"regexp"
	"strings"
)

// IgnoreInfo contains information about ignore comments in a source file
type IgnoreInfo struct {
	FileIgnored    bool
	IgnoredLines   map[int]bool    // line numbers that have disable-line comments
	IgnoredModules map[string]bool // module names that should be ignored
}

// ParseIgnoreComments analyzes source code for ignore comments
func ParseIgnoreComments(sourceCode []byte) *IgnoreInfo {
	source := string(sourceCode)
	lines := strings.Split(source, "\n")

	info := &IgnoreInfo{
		FileIgnored:    false,
		IgnoredLines:   make(map[int]bool),
		IgnoredModules: make(map[string]bool),
	}

	// Check for file-level ignore
	fileIgnorePattern := regexp.MustCompile(`//\s*nestjs-module-lint-disable-file`)
	if fileIgnorePattern.MatchString(source) {
		info.FileIgnored = true
		return info // If file is ignored, no need to check line-level ignores
	}

	// Check for line-level ignores
	lineIgnorePattern := regexp.MustCompile(`//\s*nestjs-module-lint-disable-line`)
	moduleNamePattern := regexp.MustCompile(`(\w+),?\s*//\s*nestjs-module-lint-disable-line`)

	for i, line := range lines {
		lineNum := i + 1 // Line numbers are 1-based

		if lineIgnorePattern.MatchString(line) {
			info.IgnoredLines[lineNum] = true

			// Extract module name from the line if possible
			moduleMatches := moduleNamePattern.FindStringSubmatch(line)
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
