package filesystem

import (
	"os"
)

// ReadFile reads the contents of a file
func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
