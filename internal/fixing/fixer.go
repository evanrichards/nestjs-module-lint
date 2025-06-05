package fixing

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// Fixer handles automatic fixing of unused imports
type Fixer struct {
	lang *sitter.Language
}

// NewFixer creates a new import fixer
func NewFixer(lang *sitter.Language) *Fixer {
	return &Fixer{
		lang: lang,
	}
}

// FixUnusedImports removes unused import statements and their references from module imports arrays
func (f *Fixer) FixUnusedImports(sourceCode []byte, unusedModules []string) ([]byte, error) {
	if len(unusedModules) == 0 {
		return sourceCode, nil
	}

	// Parse the source code to validate it's valid TypeScript
	tree, err := sitter.ParseCtx(context.Background(), sourceCode, f.lang)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TypeScript: %w", err)
	}

	// Check for syntax errors in the parsed tree
	if tree.HasError() {
		return nil, fmt.Errorf("TypeScript syntax error detected")
	}

	source := string(sourceCode)

	// Remove unused import statements
	for _, moduleName := range unusedModules {
		source = f.removeImportStatement(source, moduleName)
	}

	// Remove unused modules from @Module imports arrays
	source = f.removeFromModuleImports(source, unusedModules)

	// Clean up excessive blank lines
	source = f.cleanupBlankLines(source)

	return []byte(source), nil
}

// removeImportStatement removes import statements for the given module
func (f *Fixer) removeImportStatement(source, moduleName string) string {
	// Pattern to match import statements
	// Handles: import { ModuleName } from '...';
	//         import ModuleName from '...';
	//         import { ModuleName as Alias } from '...';
	patterns := []string{
		// Named import: import { ModuleName } from '...';
		fmt.Sprintf(`import\s*{\s*%s\s*}\s*from\s*['"][^'"]*['"];\s*\n?`, regexp.QuoteMeta(moduleName)),
		// Default import: import ModuleName from '...';
		fmt.Sprintf(`import\s+%s\s+from\s*['"][^'"]*['"];\s*\n?`, regexp.QuoteMeta(moduleName)),
		// Named import with alias: import { ModuleName as Alias } from '...';
		fmt.Sprintf(`import\s*{\s*[^}]*%s[^}]*}\s*from\s*['"][^'"]*['"];\s*\n?`, regexp.QuoteMeta(moduleName)),
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(source) {
			source = regex.ReplaceAllString(source, "")
			break // Remove only the first match to avoid over-removal
		}
	}

	return source
}

// removeFromModuleImports removes modules from @Module imports arrays
func (f *Fixer) removeFromModuleImports(source string, unusedModules []string) string {
	// Find @Module decorator with imports array
	modulePattern := regexp.MustCompile(`@Module\s*\(\s*{[^}]*imports\s*:\s*\[([^\]]*)\][^}]*}\s*\)`)

	return modulePattern.ReplaceAllStringFunc(source, func(moduleMatch string) string {
		// Extract the imports array content
		importsPattern := regexp.MustCompile(`imports\s*:\s*\[([^\]]*)\]`)
		importsMatch := importsPattern.FindStringSubmatch(moduleMatch)

		if len(importsMatch) < 2 {
			return moduleMatch // No imports array found, return unchanged
		}

		arrayContent := importsMatch[1]
		updatedArray := f.removeModulesFromArray(arrayContent, unusedModules)

		// Replace the imports array with the updated version
		return importsPattern.ReplaceAllString(moduleMatch, fmt.Sprintf("imports: [%s]", updatedArray))
	})
}

// removeModulesFromArray removes specified modules from an array string
func (f *Fixer) removeModulesFromArray(arrayContent string, modulesToRemove []string) string {
	// Create a set of modules to remove for efficient lookup
	removeSet := make(map[string]bool)
	for _, module := range modulesToRemove {
		removeSet[module] = true
	}

	// Check if it's a single-line or multi-line array
	if strings.Contains(arrayContent, "\n") {
		return f.removeFromMultilineArray(arrayContent, removeSet)
	} else {
		return f.removeFromInlineArray(arrayContent, removeSet)
	}
}

// removeFromInlineArray handles single-line arrays like: [A, B, C]
func (f *Fixer) removeFromInlineArray(arrayContent string, removeSet map[string]bool) string {
	// Split by comma and filter
	parts := strings.Split(arrayContent, ",")
	var kept []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		if !removeSet[trimmed] {
			kept = append(kept, part) // Keep original spacing
		}
	}

	if len(kept) == 0 {
		return "" // Empty array
	}

	return strings.Join(kept, ",")
}

// removeFromMultilineArray handles multi-line arrays with proper formatting
func (f *Fixer) removeFromMultilineArray(arrayContent string, removeSet map[string]bool) string {
	lines := strings.Split(arrayContent, "\n")
	var kept []string

	for _, line := range lines {
		moduleName := f.extractModuleName(line)
		if moduleName == "" || !removeSet[moduleName] {
			kept = append(kept, line)
		}
	}

	return strings.Join(kept, "\n")
}

// extractModuleName extracts the module name from a line in the imports array
func (f *Fixer) extractModuleName(line string) string {
	// Remove whitespace and trailing comma
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimSuffix(trimmed, ",")
	trimmed = strings.TrimSpace(trimmed)

	// Skip empty lines or comments
	if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return ""
	}

	return trimmed
}

// cleanupBlankLines removes excessive blank lines from the source
func (f *Fixer) cleanupBlankLines(source string) string {
	// Replace multiple consecutive blank lines with at most 2 blank lines
	multipleBlankLines := regexp.MustCompile(`\n\s*\n\s*\n+`)
	return multipleBlankLines.ReplaceAllString(source, "\n\n")
}
