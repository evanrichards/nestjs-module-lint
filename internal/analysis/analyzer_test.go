package analysis_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanrichards/nestjs-module-lint/internal/analysis"
)

// Mock implementations for testing

type mockModuleParser struct {
	modules   map[string]*analysis.ModuleInfo
	imports   map[string]map[string][]string
	exports   map[string]map[string][]string
	providers map[string]map[string][]string
}

func (m *mockModuleParser) ParseModuleInfo(filePath string) (*analysis.ModuleInfo, error) {
	if info, ok := m.modules[filePath]; ok {
		return info, nil
	}
	return nil, nil
}

func (m *mockModuleParser) GetImportsByModule(filePath string) (map[string][]string, error) {
	if imports, ok := m.imports[filePath]; ok {
		return imports, nil
	}
	return map[string][]string{}, nil
}

func (m *mockModuleParser) GetExportsByModule(filePath string) (map[string][]string, error) {
	if exports, ok := m.exports[filePath]; ok {
		return exports, nil
	}
	return map[string][]string{}, nil
}

func (m *mockModuleParser) GetProvidersByModule(filePath string) (map[string][]string, error) {
	if providers, ok := m.providers[filePath]; ok {
		return providers, nil
	}
	return map[string][]string{}, nil
}

func (m *mockModuleParser) GetImportPaths(filePath string) (map[string]string, error) {
	// Simple mock implementation - return empty map
	return map[string]string{}, nil
}

type mockPathResolver struct{}

func (m *mockPathResolver) ResolveImportPath(baseDir, importPath string) string {
	return filepath.Join(baseDir, importPath)
}

type mockIgnoreDetector struct{}

func (m *mockIgnoreDetector) ShouldIgnoreFile(source []byte) bool {
	// Simple mock - ignore files containing specific comment
	return string(source) == "// ignore-file"
}

func (m *mockIgnoreDetector) ShouldIgnoreImport(moduleName string, source []byte) bool {
	// Simple mock - ignore specific modules
	return moduleName == "IgnoredModule"
}

type mockReExportDetector struct{}

func (m *mockReExportDetector) GetReExportedModules(imports []string, exports []string) []string {
	var reExported []string
	exportMap := make(map[string]bool)
	for _, export := range exports {
		exportMap[export] = true
	}
	for _, imp := range imports {
		if exportMap[imp] {
			reExported = append(reExported, imp)
		}
	}
	return reExported
}

func TestAnalyzer_AnalyzeFile(t *testing.T) {
	// Create test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.module.ts")

	// Create mock parser with test data
	parser := &mockModuleParser{
		modules: map[string]*analysis.ModuleInfo{
			testFile: {
				Name:      "TestModule",
				FilePath:  testFile,
				Imports:   []string{"Module1", "Module2"},
				Exports:   []string{},
				Providers: []string{"Provider1"},
			},
		},
		imports: map[string]map[string][]string{
			testFile: {
				"TestModule": {"Module1", "Module2", "IgnoredModule"},
			},
		},
		exports: map[string]map[string][]string{
			testFile: {
				"TestModule": {"Module1"}, // Module1 is re-exported
			},
		},
		providers: map[string]map[string][]string{
			testFile: {
				"TestModule": {"Provider1"},
			},
		},
	}

	analyzer := analysis.NewAnalyzer(
		parser,
		&mockPathResolver{},
		&mockIgnoreDetector{},
		&mockReExportDetector{},
		analysis.AnalysisOptions{
			WorkingDirectory: tempDir,
			EnableIgnores:    true,
			EnableReExports:  true,
		},
	)

	// Create the test file
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	results, err := analyzer.AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// With the full dependency analysis implementation, we expect to find unused imports
	// The test setup has imports that are not used by the providers
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.ModuleName != "TestModule" {
		t.Errorf("Expected module name 'TestModule', got '%s'", result.ModuleName)
	}

	if len(result.UnusedImports) == 0 {
		t.Error("Expected to find unused imports")
	}
}

func TestAnalyzer_AnalyzeFile_IgnoredFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.module.ts")

	parser := &mockModuleParser{
		modules: map[string]*analysis.ModuleInfo{
			testFile: {
				Name:     "TestModule",
				FilePath: testFile,
			},
		},
	}

	analyzer := analysis.NewAnalyzer(
		parser,
		&mockPathResolver{},
		&mockIgnoreDetector{},
		&mockReExportDetector{},
		analysis.AnalysisOptions{
			WorkingDirectory: tempDir,
			EnableIgnores:    true,
		},
	)

	// Create file with ignore comment
	if err := os.WriteFile(testFile, []byte("// ignore-file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	results, err := analyzer.AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected no results for ignored file, got %d", len(results))
	}
}

func TestAnalyzer_AnalyzeDirectory_NoTypeScriptFiles(t *testing.T) {
	tempDir := t.TempDir()

	analyzer := analysis.NewAnalyzer(
		&mockModuleParser{},
		&mockPathResolver{},
		&mockIgnoreDetector{},
		&mockReExportDetector{},
		analysis.AnalysisOptions{
			WorkingDirectory: tempDir,
		},
	)

	_, err := analyzer.AnalyzeDirectory(tempDir)
	if err == nil {
		t.Error("Expected error for directory with no TypeScript files")
	}
}
