package app

import (
	"fmt"
	"os"
)

// FixWorkflow handles the complete fix process for a directory or file
func FixWorkflow(path string) error {
	// First, analyze to find unused imports
	reports, err := RunForDirRecursively(path)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if len(reports) == 0 {
		fmt.Println("✓ No unused imports found - nothing to fix")
		return nil
	}

	fmt.Printf("Found %d files with unused imports, fixing...\n", len(reports))

	// Fix each file
	for _, report := range reports {
		err := fixFileUnusedImports(report.Path, report.UnnecessaryImports)
		if err != nil {
			return fmt.Errorf("failed to fix %s: %w", report.Path, err)
		}
		fmt.Printf("✓ Fixed %s (removed: %v)\n", report.Path, report.UnnecessaryImports)
	}

	fmt.Printf("✓ Successfully fixed %d files\n", len(reports))
	return nil
}

// fixFileUnusedImports fixes unused imports in a specific file
func fixFileUnusedImports(filePath string, unusedModules []string) error {
	// Read the current file
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Apply fixes
	fixedCode, err := FixUnusedImports(sourceCode, unusedModules)
	if err != nil {
		return fmt.Errorf("failed to apply fixes: %w", err)
	}

	// Write the fixed code back
	err = os.WriteFile(filePath, fixedCode, 0644)
	if err != nil {
		return fmt.Errorf("failed to write fixed file: %w", err)
	}

	return nil
}
