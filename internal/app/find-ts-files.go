package app

import (
	"os"
	"path/filepath"
	"strings"
)

// FindTSFiles recursively finds all TypeScript files in the given directory.
func FindTSFiles(root string) ([]string, error) {
	var tsFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories and non-.module.ts files early
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".module.ts") {
			tsFiles = append(tsFiles, path)
		}
		return nil
	})
	return tsFiles, err
}
