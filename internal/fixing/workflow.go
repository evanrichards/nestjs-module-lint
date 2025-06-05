package fixing

import (
	"fmt"
	"os"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
)

// Workflow handles the complete fix process for a directory or file
type Workflow struct {
	analyzer analysis.ModuleAnalyzer
	fixer    *Fixer
}

// NewWorkflow creates a new fix workflow
func NewWorkflow(analyzer analysis.ModuleAnalyzer, fixer *Fixer) *Workflow {
	return &Workflow{
		analyzer: analyzer,
		fixer:    fixer,
	}
}

// FixPath handles the complete fix process for a directory or file
func (w *Workflow) FixPath(path string) error {
	// Determine if it's a file or directory and analyze accordingly
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	var reports []*analysis.ModuleAnalysisResult
	if info.IsDir() {
		reports, err = w.analyzer.AnalyzeDirectory(path)
	} else {
		reports, err = w.analyzer.AnalyzeFile(path)
	}

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
		err := w.fixFile(report.FilePath, report.UnusedImports)
		if err != nil {
			return fmt.Errorf("failed to fix %s: %w", report.FilePath, err)
		}
		fmt.Printf("✓ Fixed %s (removed: %v)\n", report.FilePath, report.UnusedImports)
	}

	fmt.Printf("✓ Successfully fixed %d files\n", len(reports))
	return nil
}

// fixFile fixes unused imports in a specific file
func (w *Workflow) fixFile(filePath string, unusedModules []string) error {
	// Read the current file
	sourceCode, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Apply fixes
	fixedCode, err := w.fixer.FixUnusedImports(sourceCode, unusedModules)
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
