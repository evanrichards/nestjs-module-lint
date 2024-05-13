package pathresolver_test

import (
	"log"
	"testing"

	pathresolver "github.com/loop-payments/nestjs-module-lint/internal/path-resolver"
)

var tsConfigFile = `
{
  "extends": "@tsconfig/node20/tsconfig.json",
  "compilerOptions": {
    "module": "CommonJS",
    "moduleResolution": "node10",
    "lib": ["es2023", "esnext.disposable"],
    "declaration": true,
    "removeComments": true,
    "emitDecoratorMetadata": true,
    "experimentalDecorators": true,
    "allowSyntheticDefaultImports": true,
    "sourceMap": true,
    "outDir": "./dist/src",
    "baseUrl": "./",
    "jsx": "react-jsx",
    "strict": true,
    "incremental": true,
    "skipLibCheck": true,
    "noImplicitAny": true,
    // make DTO classes easier to write
    "strictPropertyInitialization": false,
    "forceConsistentCasingInFileNames": true,
    "noFallthroughCasesInSwitch": true,
    "useDefineForClassFields": false,
    "paths": {
      "src/*": ["./src/*"],
      "@testing": ["./test/for/stuff"]
    },
    "resolveJsonModule": true
  },
  "references": [
    { "path": "lib/email-templates" },
    { "path": "lib/csv-templates" },
    { "path": "lib/pdf-templates" }
  ],
  "ts-node": {
    "swc": true
  },
  "watchOptions": {
    "excludeFiles": ["src/common/util/test-util/*"]
  }
}

`

func TestParseTsConfig(t *testing.T) {
	// Example TypeScript source to parse
	parsed, err := pathresolver.ParseTsConfigFile([]byte(tsConfigFile))
	if err != nil {
		log.Fatalf("Failed to parse tsconfig file: %v", err)
	}
	if parsed == nil {
		t.Fatalf("Failed to parse tsconfig file")
	}
	paths := parsed.CompilerOptions.Paths
	if len(paths) != 2 {
		t.Fatalf("Failed to parse paths")
	}
	if len(paths["src/*"]) != 1 {
		t.Fatalf("Failed to parse paths")
	}
	if len(paths["@testing"]) != 1 {
		t.Fatalf("Failed to parse paths")
	}

}

func TestTsPathResovler(t *testing.T) {
	tsPathResolver, err := pathresolver.NewTsPathResovler(
		[]byte(tsConfigFile), "tsconfig.json", "/path/to/project")
	if err != nil {
		log.Fatalf("Failed to create ts path resolver: %v", err)
	}
	t.Run("resolve @paths", func(t *testing.T) {
		path := tsPathResolver.ResolveImportPath("@testing/my-new-test")
		if path != "/path/to/project/test/for/stuff/my-new-test" {
			t.Fatalf("Failed to resolve path: %v", path)
		}
	})
	t.Run("resolve local paths", func(t *testing.T) {
		path := tsPathResolver.ResolveImportPath("./src/my-new-test")
		if path != "/path/to/project/src/my-new-test" {
			t.Fatalf("Failed to resolve path: %v", path)
		}
	})
}
