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

	// Check if there's a blank line after imports before we start modifying
	hasBlankLineAfterImports := f.hasBlankLineAfterImports(source)

	// Remove unused import statements
	for _, moduleName := range unusedModules {
		source = f.removeImportStatement(source, moduleName)
	}

	// Remove unused modules from @Module imports arrays
	source = f.removeFromModuleImports(source, unusedModules)

	// If there was a blank line after imports originally, ensure it's preserved
	if hasBlankLineAfterImports {
		source = f.ensureBlankLineAfterImports(source)
	}

	// Clean up excessive blank lines (but preserve intended single blank lines)
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
		// Named import: import { ModuleName } from '...'; [optional comment]
		fmt.Sprintf(`import\s*{\s*%s\s*}\s*from\s*['"][^'"]*['"];[^\n]*\n?`, regexp.QuoteMeta(moduleName)),
		// Default import: import ModuleName from '...'; [optional comment]
		fmt.Sprintf(`import\s+%s\s+from\s*['"][^'"]*['"];[^\n]*\n?`, regexp.QuoteMeta(moduleName)),
		// Named import with alias: import { ModuleName as Alias } from '...'; [optional comment]
		fmt.Sprintf(`import\s*{\s*[^}]*%s[^}]*}\s*from\s*['"][^'"]*['"];[^\n]*\n?`, regexp.QuoteMeta(moduleName)),
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
			kept = append(kept, trimmed) // Use trimmed to normalize spacing
		}
	}

	if len(kept) == 0 {
		return "" // Empty array
	}

	// Join with proper spacing: comma followed by space
	return strings.Join(kept, ", ")
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

	// Skip empty lines or comment-only lines
	if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return ""
	}

	// Remove inline comments (e.g., "ModuleName // comment" -> "ModuleName")
	if commentIndex := strings.Index(trimmed, "//"); commentIndex != -1 {
		trimmed = strings.TrimSpace(trimmed[:commentIndex])
	}
	if commentIndex := strings.Index(trimmed, "/*"); commentIndex != -1 {
		trimmed = strings.TrimSpace(trimmed[:commentIndex])
	}

	// Remove trailing comma again (in case it was after the module name but before the comment)
	trimmed = strings.TrimSuffix(trimmed, ",")
	trimmed = strings.TrimSpace(trimmed)

	return trimmed
}

// hasBlankLineAfterImports checks if there's a blank line between imports and @Module
func (f *Fixer) hasBlankLineAfterImports(source string) bool {
	// Look for the pattern: import statements followed by blank line(s) followed by @Module
	pattern := regexp.MustCompile(`import\s+[^;]+;.*?\n\s*\n\s*@Module`)
	return pattern.MatchString(source)
}

// ensureBlankLineAfterImports ensures there's a blank line between imports and @Module
func (f *Fixer) ensureBlankLineAfterImports(source string) string {
	// Pattern to match: (last import)(possible whitespace/newlines)(@Module)
	pattern := regexp.MustCompile(`(import\s+[^;]+;\s*\n)(\s*)(@Module)`)

	return pattern.ReplaceAllStringFunc(source, func(match string) string {
		// Extract the components
		submatches := pattern.FindStringSubmatch(match)
		if len(submatches) != 4 {
			return match
		}

		lastImport := submatches[1]
		whitespace := submatches[2]
		moduleDecorator := submatches[3]

		// If there's no blank line (just whitespace, no empty lines), add one
		if !strings.Contains(whitespace, "\n\n") && !strings.Contains(whitespace, "\n\r\n") && !strings.Contains(whitespace, "\r\n\r\n") {
			return lastImport + "\n" + moduleDecorator
		}

		return match
	})
}

// cleanupBlankLines removes excessive blank lines from the source while preserving single blank lines
func (f *Fixer) cleanupBlankLines(source string) string {
	// Replace 3 or more consecutive blank lines with exactly 2 blank lines
	// This preserves single blank lines (which are often intentional formatting)
	excessiveBlankLines := regexp.MustCompile(`(\n\s*){3,}`)
	return excessiveBlankLines.ReplaceAllString(source, "\n\n")
}
