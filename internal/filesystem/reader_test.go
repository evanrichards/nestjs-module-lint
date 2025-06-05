package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/filesystem"
)

func TestReadFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.ts")
	testContent := []byte(`import { Module } from '@nestjs/common';

@Module({
  imports: [],
  providers: [],
})
export class TestModule {}`)

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading the file
	content, err := filesystem.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("Read content does not match expected content")
	}
}

func TestReadFile_NonExistent(t *testing.T) {
	_, err := filesystem.ReadFile("/non/existent/file.ts")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.ts")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !filesystem.FileExists(testFile) {
		t.Error("Expected FileExists to return true for existing file")
	}

	// Test non-existent file
	if filesystem.FileExists(filepath.Join(tempDir, "nonexistent.ts")) {
		t.Error("Expected FileExists to return false for non-existent file")
	}
}

func TestIsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.ts")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test directory
	if !filesystem.IsDirectory(tempDir) {
		t.Error("Expected IsDirectory to return true for directory")
	}

	// Test file
	if filesystem.IsDirectory(testFile) {
		t.Error("Expected IsDirectory to return false for file")
	}

	// Test non-existent path
	if filesystem.IsDirectory("/non/existent/path") {
		t.Error("Expected IsDirectory to return false for non-existent path")
	}
}
