package filesystem

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrNoTypeScriptFiles is returned when no TypeScript files are found
var ErrNoTypeScriptFiles = errors.New("no TypeScript files found in directory")

// FindTypeScriptFiles recursively finds all TypeScript files in a directory
func FindTypeScriptFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isTypeScriptFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, ErrNoTypeScriptFiles
	}

	return files, nil
}

// isTypeScriptFile checks if a file has a TypeScript extension
func isTypeScriptFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".ts") || strings.HasSuffix(lower, ".tsx")
}
