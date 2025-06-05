package app

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// FixUnusedImports removes unused import statements and their references from module imports arrays
func FixUnusedImports(sourceCode []byte, unusedModules []string) ([]byte, error) {
	if len(unusedModules) == 0 {
		return sourceCode, nil
	}

	// Parse the source code to validate it's valid TypeScript
	tree, err := sitter.ParseCtx(context.Background(), sourceCode, lang)
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
		source = removeImportStatement(source, moduleName)
	}

	// Remove unused modules from @Module imports arrays
	source = removeFromModuleImports(source, unusedModules)

	// Clean up excessive blank lines (more than one consecutive blank line)
	source = cleanupBlankLines(source)

	return []byte(source), nil
}

// removeImportStatement removes the import statement for the given module
func removeImportStatement(source, moduleName string) string {
	// Pattern to match import statements:
	// import { ModuleName } from "...";
	// import ModuleName from "...";
	// import { ModuleName as Alias } from "...";
	// import { SomeModule as ModuleName } from "...";

	patterns := []string{
		// Named import: import { ModuleName } from "...";
		fmt.Sprintf(`import\s*{\s*%s\s*}\s*from\s*["`+"`"+`'].*?["`+"`"+`'].*;?\r?\n`, regexp.QuoteMeta(moduleName)),
		// Default import: import ModuleName from "...";
		fmt.Sprintf(`import\s+%s\s+from\s*["`+"`"+`'].*?["`+"`"+`'].*;?\r?\n`, regexp.QuoteMeta(moduleName)),
		// Named import with alias: import { SomeModule as ModuleName } from "...";
		fmt.Sprintf(`import\s*{\s*.*?\s+as\s+%s\s*}\s*from\s*["`+"`"+`'].*?["`+"`"+`'].*;?\r?\n`, regexp.QuoteMeta(moduleName)),
		// Named import where ModuleName is aliased: import { ModuleName as SomeAlias } from "...";
		fmt.Sprintf(`import\s*{\s*%s\s+as\s+.*?\s*}\s*from\s*["`+"`"+`'].*?["`+"`"+`'].*;?\r?\n`, regexp.QuoteMeta(moduleName)),
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(source) {
			// Replace with just a newline to maintain spacing
			source = re.ReplaceAllString(source, "\n")
			break
		}
	}

	return source
}

// removeFromModuleImports removes modules from the imports array in @Module decorators
func removeFromModuleImports(source string, unusedModules []string) string {
	// Find @Module({ imports: [...] }) patterns
	modulePattern := regexp.MustCompile(`(@Module\s*\(\s*{\s*[^}]*imports\s*:\s*\[)([^\]]*?)(\][^}]*}\s*\))`)

	return modulePattern.ReplaceAllStringFunc(source, func(match string) string {
		parts := modulePattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}

		prefix := parts[1]       // @Module({ imports: [
		importsArray := parts[2] // content inside the array
		suffix := parts[3]       // ] })

		// Clean up the imports array
		cleanedArray := removeModulesFromArray(importsArray, unusedModules)

		return prefix + cleanedArray + suffix
	})
}

// removeModulesFromArray removes specified modules from an array string
func removeModulesFromArray(arrayContent string, modulesToRemove []string) string {
	if strings.TrimSpace(arrayContent) == "" {
		return arrayContent
	}

	// Create a set for O(1) lookup
	removeSet := make(map[string]bool)
	for _, module := range modulesToRemove {
		removeSet[module] = true
	}

	// Check if this is a multiline array (contains newlines)
	isMultiline := strings.Contains(arrayContent, "\n")

	if isMultiline {
		return removeFromMultilineArray(arrayContent, removeSet)
	} else {
		return removeFromInlineArray(arrayContent, removeSet)
	}
}

// removeFromInlineArray handles inline arrays like [ModuleA, ModuleB, ModuleC]
func removeFromInlineArray(arrayContent string, removeSet map[string]bool) string {
	// Split by comma and filter
	parts := strings.Split(arrayContent, ",")
	var kept []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		// Extract module name (handle comments)
		moduleName := extractModuleName(trimmed)
		if moduleName != "" && !removeSet[moduleName] {
			kept = append(kept, strings.TrimSpace(part))
		}
	}

	if len(kept) == 0 {
		return ""
	}

	return strings.Join(kept, ", ")
}

// removeFromMultilineArray handles multiline arrays with proper indentation
func removeFromMultilineArray(arrayContent string, removeSet map[string]bool) string {
	lines := strings.Split(arrayContent, "\n")
	var kept []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			kept = append(kept, line) // Keep empty lines
			continue
		}

		// Extract module name from the line
		moduleName := extractModuleName(trimmed)
		if moduleName == "" || !removeSet[moduleName] {
			kept = append(kept, line)
		}
	}

	return strings.Join(kept, "\n")
}

// extractModuleName extracts the module name from a line, handling comments and trailing commas
func extractModuleName(line string) string {
	// Remove inline comments
	if idx := strings.Index(line, "//"); idx != -1 {
		line = line[:idx]
	}

	// Remove trailing comma and whitespace
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ",")
	line = strings.TrimSpace(line)

	// The remaining should be the module name
	return line
}

// cleanupBlankLines removes excessive consecutive blank lines, keeping at most one
func cleanupBlankLines(source string) string {
	// Replace multiple consecutive newlines with just two newlines (one blank line)
	re := regexp.MustCompile(`\n\s*\n\s*\n+`)
	source = re.ReplaceAllString(source, "\n\n")

	// Special handling for import section - remove blank lines between imports
	lines := strings.Split(source, "\n")
	var result []string
	var inImportSection bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect if we're in an import section
		if strings.HasPrefix(trimmed, "import ") {
			inImportSection = true
			result = append(result, line)
		} else if inImportSection && trimmed == "" {
			// Check if next non-empty line is still an import
			nextIsImport := false
			for j := i + 1; j < len(lines); j++ {
				nextTrimmed := strings.TrimSpace(lines[j])
				if nextTrimmed != "" {
					nextIsImport = strings.HasPrefix(nextTrimmed, "import ")
					break
				}
			}
			if !nextIsImport {
				// We're leaving the import section
				inImportSection = false
				result = append(result, line)
			}
			// Skip blank lines within import section
		} else {
			if inImportSection && trimmed != "" {
				inImportSection = false
			}
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
