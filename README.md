# NestJS Module Lint

[![npm version](https://badge.fury.io/js/nestjs-module-lint.svg)](https://badge.fury.io/js/nestjs-module-lint)
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

If you have Go 1.22+ installed:

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

### Command Options

```bash
nestjs-module-lint import-lint [flags] <path>

Flags:
  -h, --help   help for import-lint
      --json   Output in JSON format
      --text   Output in text format (default)
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
    "lint:modules:json": "nestjs-module-lint import-lint --json src/"
  }
}
```

### CI/CD Integration

**GitHub Actions:**

```yaml
name: Module Lint
on: [push, pull_request]

jobs:
  module-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm ci
      - run: npm run lint:modules
```

**Pre-commit Hook:**

```json
{
  "husky": {
    "hooks": {
      "pre-commit": "nestjs-module-lint import-lint src/"
    }
  }
}
```

## 🛠️ Development

### Building from Source

```bash
git clone https://github.com/evanrichards/nestjs-module-lint.git
cd nestjs-module-lint
go mod download
go build -o nestjs-module-lint .
```

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Commit your changes: `git commit -am 'Add my feature'`
6. Push to the branch: `git push origin feature/my-feature`
7. Submit a pull request

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