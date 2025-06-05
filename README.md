# NestJS Module Lint

[![npm version](https://badge.fury.io/js/nestjs-module-lint.svg)](https://badge.fury.io/js/nestjs-module-lint)
[![CI](https://github.com/evanrichards/nestjs-module-lint/actions/workflows/ci.yml/badge.svg)](https://github.com/evanrichards/nestjs-module-lint/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful command-line tool for analyzing NestJS modules to identify unused module imports in `@Module()` decorators. Detects when modules are imported but their exports are never used by the module's providers or controllers. Built with Go and tree-sitter for fast and accurate TypeScript parsing.

## 🚀 Features

- **Fast Analysis**: Built with Go and tree-sitter for high-performance TypeScript parsing
- **Unused Module Detection**: Identifies modules in `@Module()` imports arrays that aren't actually used
- **Multiple Output Formats**: Support for both text and JSON output
- **Recursive Directory Scanning**: Analyze entire project directories or individual files
- **CI/CD Integration**: Perfect for automated code quality checks
- **Cross-Platform**: Works on macOS, Linux, and Windows

## 📦 Installation

### NPM/Yarn (Recommended)

Install as a development dependency:

```bash
npm install --save-dev nestjs-module-lint
```

Or with Yarn:

```bash
yarn add --dev nestjs-module-lint
```

For global installation:

```bash
npm install -g nestjs-module-lint
```

### Go Install (Alternative)

If you have Go 1.21+ installed:

```bash
go install github.com/evanrichards/nestjs-module-lint@latest
```

### Manual Build

```bash
git clone https://github.com/evanrichards/nestjs-module-lint.git
cd nestjs-module-lint
go build -o nestjs-module-lint .
```

## 🔧 Usage

### Basic Usage

Analyze a single module file:

```bash
npx nestjs-module-lint import-lint src/app/app.module.ts
```

Analyze an entire directory recursively:

```bash
npx nestjs-module-lint import-lint src/
```

### Output Formats

**Text Output (Default):**
```bash
npx nestjs-module-lint import-lint src/app/app.module.ts
```

**JSON Output:**
```bash
npx nestjs-module-lint import-lint --json src/app/app.module.ts
```

### Auto-Fix Unused Imports

**Preview Changes:**
```bash
npx nestjs-module-lint import-lint src/app/app.module.ts
```

**Automatically Fix:**
```bash
npx nestjs-module-lint import-lint --fix src/app/app.module.ts
```

The `--fix` flag will:
- Remove unused import statements from the top of files
- Clean up the `imports: [...]` arrays in `@Module()` decorators  
- Preserve formatting and handle both inline and multiline arrays
- Support all import types: named, default, and aliased imports

### Command Options

```bash
nestjs-module-lint import-lint [flags] <path>

Output Flags:
      --json        Output in JSON format
      --text        Output in text format (default)

Fix Flags:
      --fix         Automatically remove unused imports

CI/CD Flags:
      --check       Check mode with pass/fail output (good for CI)
      --exit-zero   Exit with code 0 even when issues are found
      --quiet       Suppress output (useful with --exit-zero)

Other:
  -h, --help        help for import-lint
```

## 📋 Prerequisites

- **Node.js**: Version 14.0 or higher
- **TypeScript Project**: Must have a `tsconfig.json` file in your project root
- **NestJS**: Compatible with NestJS projects using standard module patterns

## 📖 How It Works

The tool analyzes your NestJS modules by:

1. **Parsing TypeScript**: Uses tree-sitter to build an Abstract Syntax Tree (AST) of your TypeScript files
2. **Module Analysis**: Identifies `@Module()` decorators and extracts their imports, providers, controllers, and exports arrays
3. **Dependency Tracking**: For each module in the imports array, checks if any of its exports are used by the current module's providers or controllers
4. **Unused Detection**: Reports modules in the imports array whose exports are never actually used

### Example Analysis

Given this NestJS module:

```typescript
import { Module } from '@nestjs/common';
import { UsersService } from './users.service';
import { AuthModule } from '../auth/auth.module';
import { EmailModule } from '../email/email.module';
import { LoggingModule } from '../logging/logging.module';

@Module({
  imports: [
    AuthModule,     // Used: UsersService uses AuthService from AuthModule
    EmailModule,    // UNUSED: No provider uses EmailService from EmailModule
    LoggingModule,  // UNUSED: No provider uses LoggingService from LoggingModule
  ],
  providers: [UsersService],
  exports: [UsersService],
})
export class UsersModule {}
```

If `UsersService` only injects `AuthService` but never uses exports from `EmailModule` or `LoggingModule`, the tool will report:

```
Module: UsersModule
Path: src/users/users.module.ts
Unnecessary Imports:
	EmailModule
	LoggingModule

Total number of modules with unused imports: 1
```

## 🔄 Integration

### Package.json Scripts

Add to your `package.json`:

```json
{
  "scripts": {
    "lint:modules": "nestjs-module-lint import-lint src/",
    "lint:modules:fix": "nestjs-module-lint import-lint --fix src/",
    "lint:modules:json": "nestjs-module-lint import-lint --json src/"
  }
}
```

### CI/CD Integration

The tool is designed with CI/CD best practices in mind:

**Exit Codes:**
- `0` - No unused imports found (or `--exit-zero` flag used)
- `1` - Unused imports found (fails CI/CD pipeline)
- `2` - Execution error (invalid path, parsing errors, etc.)

**CI/CD Modes:**

```bash
# Standard mode (exit 1 if issues found)
nestjs-module-lint import-lint src/

# Check mode (clear pass/fail for CI)
nestjs-module-lint import-lint --check src/

# Report mode (never fail CI, just report)
nestjs-module-lint import-lint --exit-zero --quiet src/
```

**GitHub Actions:**

```yaml
name: Module Lint
on: [push, pull_request]

jobs:
  module-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm ci
      - run: npx nestjs-module-lint import-lint --check src/
```

**Pre-commit Hook:**

```json
{
  "husky": {
    "hooks": {
      "pre-commit": "npx nestjs-module-lint import-lint --check src/"
    }
  }
}
```

## 🛠️ Development

This project includes a comprehensive Makefile for easy development and CI/CD integration.

### Quick Start

```bash
git clone https://github.com/evanrichards/nestjs-module-lint.git
cd nestjs-module-lint
make help    # See all available commands
make build   # Build the binary
make test    # Run tests
```

### Available Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary to `bin/nestjs-module-lint` |
| `make test` | Run all tests with verbose output |
| `make bench` | Run benchmarks on core packages |
| `make lint` | Run golangci-lint for code quality |
| `make fmt` | Format all Go code |
| `make clean` | Remove build artifacts and clear cache |
| `make install` | Install binary to `$GOPATH/bin` |
| `make run` | Build and run with `test.ts` |
| `make run-json` | Build and run with JSON output |
| `make check` | Build and run in check mode (CI-friendly) |
| `make help` | Show all available targets |

### Manual Build

```bash
go mod download
go build -o bin/nestjs-module-lint .
```

### Quality Assurance

The project uses automated CI/CD with comprehensive testing:

- **Multi-OS Testing**: Ubuntu, Windows, macOS
- **Multi-Go Version**: Go 1.21, 1.22
- **Automated Linting**: golangci-lint with latest rules
- **Benchmark Testing**: Performance regression detection

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Run the development workflow:
   ```bash
   make fmt      # Format code
   make lint     # Check code quality
   make test     # Run tests
   make bench    # Run benchmarks
   ```
5. Commit your changes: `git commit -am 'Add my feature'`
6. Push to the branch: `git push origin feature/my-feature`
7. Submit a pull request

The CI pipeline will automatically run tests across multiple platforms and Go versions.

## 📊 Output Examples

### Text Output
```
Module: AppModule
Path: src/app/app.module.ts
Unnecessary Imports:
	EmailModule
	LoggingModule

Module: UsersModule  
Path: src/users/users.module.ts
Unnecessary Imports:
	NotificationModule

Total number of modules with unused imports: 2
```

### JSON Output
```json
[
  {
    "module_name": "AppModule",
    "path": "src/app/app.module.ts",
    "unnecessary_imports": ["EmailModule", "LoggingModule"]
  },
  {
    "module_name": "UsersModule",
    "path": "src/users/users.module.ts", 
    "unnecessary_imports": ["NotificationModule"]
  }
]
```

## 🗺️ Features & Roadmap

### ✅ Current Features
- **Import Analysis**: Detect unused module imports in `@Module()` decorators
- **Auto-Fix Capability**: Automatically remove unused imports with `--fix` flag
- **Multiple Output Formats**: Text and JSON output support
- **CI/CD Integration**: Standardized exit codes and check modes
- **Cross-Platform**: Works on macOS, Linux, Windows
- **TypeScript Path Mapping**: Full support for tsconfig.json paths
- **Performance Optimized**: Built with Go and tree-sitter for speed

### 🚧 Planned Features
  
#### Export Analysis
- **`export-lint` Command**: Find unused exports in NestJS modules
  ```bash
  # Find exports that are never imported by other modules
  nestjs-module-lint export-lint src/
  
  # Combined import + export analysis
  nestjs-module-lint lint src/
  ```

#### Ignore Comments
- **File-level Ignores**: Skip analysis for specific files
  ```typescript
  // nestjs-module-lint-disable-file
  @Module({
    imports: [SomeModuleWeWantToKeep], // This won't be flagged
  })
  export class LegacyModule {}
  ```
  
- **Line-level Ignores**: Skip specific imports
  ```typescript
  @Module({
    imports: [
      RequiredModule,
      OptionalModule, // nestjs-module-lint-disable-line
    ],
  })
  export class MyModule {}
  ```

#### Project-Level Configuration
- **Configuration File**: `.nestjs-module-lint.json` or `nestjs-module-lint.config.js` for project-wide settings
  ```json
  {
    "ignoreModules": ["LegacyModule", "ThirdPartyModule"],
    "ignoreDirectories": ["src/legacy/**", "src/external/**"],
    "ignoreSubdomains": ["@company/legacy", "@deprecated/*"],
    "rules": {
      "allowUnusedInTests": true,
      "strictMode": false
    },
    "exclude": ["**/*.spec.ts", "**/*.test.ts"]
  }
  ```
  - **Module Allowlisting**: Never error on specific modules by name
  - **Directory Exclusion**: Skip entire subdirectories from analysis
  - **Subdomain Ignoring**: Ignore imports from specific npm scopes or patterns
  - **Rule Customization**: Fine-tune linting behavior per project
  - **Test File Handling**: Special rules for test files and mocks

#### Advanced Analysis
- **Dependency Graph**: Visualize module dependencies
- **Circular Dependency Detection**: Find circular imports between modules
- **Dead Code Analysis**: Find modules that are never imported anywhere
- **Module Health Score**: Overall module dependency health metrics

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🐛 Issues & Support

If you encounter any issues or have questions:

- 🐛 [Report bugs](https://github.com/evanrichards/nestjs-module-lint/issues)
- 💡 [Request features](https://github.com/evanrichards/nestjs-module-lint/issues)
- 📖 [Check documentation](https://github.com/evanrichards/nestjs-module-lint)

## 🏷️ Changelog

See [Releases](https://github.com/evanrichards/nestjs-module-lint/releases) for version history and changes.
