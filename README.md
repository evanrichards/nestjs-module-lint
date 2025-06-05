# NestJS Module Linter

NestJS Module Linter is a command-line tool designed to help developers
identify unused exports and imports in NestJS modules. This tool uses the
`go-tree-sitter` library to parse TypeScript files and analyze the dependencies
between modules within a NestJS project. By highlighting unused dependencies,
it helps maintain a clean and efficient codebase.

## Features

- Analyze TypeScript files to find unused imports and exports in NestJS modules.
- CLI-based output for easy integration with development workflows and CI/CD
  pipelines.

## Getting Started

### Prerequisites

- Node.js 14.0 or higher

### Installation

Install using npm:

```bash
npm install --save-dev nestjs-module-lint
```

Or using Yarn:

```bash
yarn add --dev nestjs-module-lint
```

For global installation:

```bash
npm install -g nestjs-module-lint
```

### Alternative Installation Methods

If you have Go installed, you can also install directly from source:

```bash
go install github.com/loop-payments/nestjs-module-lint@latest
```

Or clone and build manually:

```bash
git clone https://github.com/loop-payments/nestjs-module-lint.git
cd nestjs-module-lint
go build -o nestjs-module-lint .
```

## Usage

Run the linter by specifying the path to a NestJS module file:

```bash
npx nestjs-module-lint <path-to-module>
```

Or with Yarn:

```bash
yarn nestjs-module-lint <path-to-module>
```

If installed globally:

```bash
nestjs-module-lint <path-to-module>
```

### Example

```bash
npx nestjs-module-lint src/app/app.module.ts
```

This command will analyze the specified module and report any unused imports
or exports detected.

## How It Works

The tool parses the specified TypeScript module files to build an abstract
syntax tree (AST) using go-tree-sitter. It then recursively analyzes all
imports and their corresponding exports in the file and other related module
files. The output specifies whether an imported module's exports are used in
the importing module.

## Contributing

Contributions are welcome! Please feel free to submit pull requests, report
bugs, or suggest features.

## Development

After cloning the repository, navigate to the project directory and make
modifications as needed. Use Go's built-in tooling for testing:

```bash
go test ./...
```

## License

Distributed under the MIT License. See LICENSE for more information.
