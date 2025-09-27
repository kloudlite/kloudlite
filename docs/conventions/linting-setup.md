# Linting Setup Guide

This guide explains the linting configuration for Kloudlite v2 and how to use it effectively.

## 🔧 Overview

The project uses different linters for different parts of the codebase:

- **Backend (Go)**: golangci-lint
- **Frontend (TypeScript/React)**: ESLint + Prettier
- **General**: EditorConfig + Pre-commit hooks

## 📦 Installation

### Prerequisites

```bash
# Install pre-commit
pip install pre-commit

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Frontend dependencies (already in package.json)
cd web && pnpm install
```

### Setup Pre-commit Hooks

```bash
# Install pre-commit hooks
pre-commit install

# Run on all files (first time)
pre-commit run --all-files
```

## 🔍 Configuration Files

### 1. ESLint Configuration (`web/.eslintrc.json`)

- Extends Next.js and TypeScript recommended rules
- Enforces import ordering
- Prevents console.log and debugger statements
- Configures TypeScript-specific rules

Key rules:
- No unused variables (except those prefixed with `_`)
- Warn on `any` types
- Enforce React hooks rules
- Single quotes, no semicolons
- Trailing commas in multi-line

### 2. Prettier Configuration (`web/.prettierrc`)

- No semicolons
- Single quotes
- 100 character line width
- Trailing commas
- Automatic import sorting

### 3. Golangci-lint Configuration (`api/.golangci.yml`)

Enabled linters:
- Standard Go linters (vet, errcheck, etc.)
- Security (gosec)
- Code complexity (gocyclo, gocognit)
- Style (gofmt, goimports, stylecheck)
- Performance (prealloc)

Key settings:
- 5-minute timeout
- Skip generated files (*.pb.go)
- Cognitive complexity: 15
- Cyclomatic complexity: 15

### 4. EditorConfig (`.editorconfig`)

Ensures consistent formatting across editors:
- UTF-8 encoding
- LF line endings
- Trim trailing whitespace
- Language-specific indentation

### 5. Pre-commit Hooks (`.pre-commit-config.yaml`)

Runs automatically before commits:
- Trailing whitespace removal
- File ending fixes
- Large file detection
- Private key detection
- Go formatting and imports
- ESLint and Prettier
- TypeScript type checking
- Proto file linting
- Security scanning

## 📋 Usage

### Running Linters Manually

#### Frontend
```bash
cd web

# Run ESLint
pnpm lint

# Fix ESLint issues
pnpm lint:fix

# Run Prettier
pnpm prettier --check .

# Fix Prettier issues
pnpm prettier --write .

# Type checking
pnpm typecheck
```

#### Backend
```bash
cd api

# Run golangci-lint
golangci-lint run

# Run on specific directory
golangci-lint run ./apps/auth/...

# Auto-fix issues (where possible)
golangci-lint run --fix
```

### IDE Integration

#### VS Code

Install extensions:
- ESLint
- Prettier
- Go
- EditorConfig

Add to `.vscode/settings.json`:
```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package"
}
```

#### JetBrains IDEs

1. Enable EditorConfig support
2. Configure ESLint: Settings → Languages → JavaScript → Code Quality Tools
3. Configure Prettier: Settings → Languages → JavaScript → Prettier
4. Configure golangci-lint: Settings → Go → Go Linter

## 🚨 Common Issues and Fixes

### ESLint Issues

```typescript
// ❌ Unused variable
const unused = 'value'

// ✅ Prefix with underscore if intentionally unused
const _unused = 'value'

// ❌ Console.log
console.log('debug')

// ✅ Use proper logging
logger.debug('debug')

// ❌ Any type
let value: any

// ✅ Use specific type
let value: string | number
```

### Go Issues

```go
// ❌ Unused import
import (
    "fmt"
    "unused/package"
)

// ✅ Remove unused imports or use goimports

// ❌ Error not checked
result, _ := someFunction()

// ✅ Always check errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed: %w", err)
}

// ❌ High complexity
func complexFunction() {
    // 20+ lines of nested if/else
}

// ✅ Break into smaller functions
func simpleFunction() {
    step1()
    step2()
}
```

## 🎯 Best Practices

1. **Run linters before committing** - Pre-commit hooks help, but run manually for faster feedback
2. **Fix issues immediately** - Don't let linting errors accumulate
3. **Configure your IDE** - Real-time feedback prevents issues
4. **Don't disable rules globally** - Use inline comments for specific exceptions
5. **Keep configurations updated** - Review and update linting rules periodically

## 📝 Disabling Rules (When Necessary)

### TypeScript/ESLint
```typescript
// Disable for next line
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const value: any = getData()

// Disable for file
/* eslint-disable no-console */
```

### Go
```go
// Disable for next line
//nolint:errcheck
result, _ := someFunction()

// Disable for function
//nolint:gocyclo
func necessarilyComplexFunction() {
    // Complex logic
}
```

### Prettier
```typescript
// prettier-ignore
const matrix = [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9]
]
```

## 🔄 Updating Configurations

When updating linting configurations:

1. Discuss changes with the team
2. Run on entire codebase to assess impact
3. Fix issues or adjust rules as needed
4. Update this documentation
5. Notify team of changes

Remember: Linting helps maintain code quality and consistency. Embrace it!