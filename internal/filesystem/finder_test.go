package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/filesystem"
)

func TestFindTypeScriptFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"module1.ts",
		"module2.tsx",
		"component.ts",
		"nested/module3.ts",
		"nested/deep/module4.tsx",
		"not-typescript.js",
		"README.md",
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte("// test file"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Test finding TypeScript files
	files, err := filesystem.FindTypeScriptFiles(tempDir)
	if err != nil {
		t.Fatalf("FindTypeScriptFiles failed: %v", err)
	}

	// Check that we found the right number of TypeScript files
	expectedCount := 5 // .ts and .tsx files only
	if len(files) != expectedCount {
		t.Errorf("Expected %d TypeScript files, got %d", expectedCount, len(files))
	}

	// Verify all returned files have TypeScript extensions
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext != ".ts" && ext != ".tsx" {
			t.Errorf("Found non-TypeScript file: %s", file)
		}
	}
}

func TestFindTypeScriptFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	files, err := filesystem.FindTypeScriptFiles(tempDir)
	if err != filesystem.ErrNoTypeScriptFiles {
		t.Errorf("Expected ErrNoTypeScriptFiles, got %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected no files, got %d", len(files))
	}
}

func TestFindTypeScriptFiles_NonExistentDirectory(t *testing.T) {
	_, err := filesystem.FindTypeScriptFiles("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}
