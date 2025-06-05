# CI Test Files

This directory contains test files used by the CI/CD pipeline and development workflows.

## Files

- `test.module.ts` - A minimal NestJS module used for testing the CLI functionality

## Important

**DO NOT DELETE** these files. They are required for:

1. CI testing in GitHub Actions
2. Local development with `make check`, `make run`, etc.
3. Pre-commit hooks that validate the tool functionality

The pre-commit hook will prevent commits if these files are missing.

## Usage

The CLI analyzes `test.module.ts` to verify that:
- The tool can parse TypeScript files correctly
- Module detection works properly 
- Import analysis functions as expected

This file intentionally has no unused imports to ensure CI passes consistently.