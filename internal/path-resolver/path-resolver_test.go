package pathresolver_test

import (
	"testing"

	pathresolver "github.com/loop-payments/nestjs-module-lint/internal/path-resolver"
)

func TestResolveImportPath(t *testing.T) {
	tsConfig := `{
		"compilerOptions": {
			"paths": {
				"src/*": ["./src/*"],
				"@testing/*": ["./test/for/stuff/*"]
			}
		}
	}`

	tests := []struct {
		name       string
		importPath string
		expected   string
	}{
		{"resolve @paths", "@testing/my-new-test", "/path/to/project/test/for/stuff/my-new-test.ts"},
		{"resolve local paths", "./my-new-test", "/unit-test/my-new-test.ts"},
		{"resolve package local paths", "src/my-new-test", "/path/to/project/src/my-new-test.ts"},
	}

	tsPathResolver, err := pathresolver.NewTsPathResolver([]byte(tsConfig), "/path/to/project")
	if err != nil {
		t.Fatalf("Failed to create ts path resolver: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tsPathResolver.ResolveImportPath("/unit-test", tt.importPath); got != tt.expected {
				t.Errorf("ResolveImportPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}
