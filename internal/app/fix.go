package app

import (
	"github.com/evanrichards/nestjs-module-lint/internal/fixing"
)

// FixUnusedImports removes unused import statements and their references from module imports arrays
func FixUnusedImports(sourceCode []byte, unusedModules []string) ([]byte, error) {
	fixer := fixing.NewFixer(getTypescriptLanguage())
	return fixer.FixUnusedImports(sourceCode, unusedModules)
}
