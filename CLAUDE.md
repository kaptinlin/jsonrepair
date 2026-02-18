# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with the jsonrepair package.

## Project Overview

**jsonrepair** is a Go port of the [jsonrepair JavaScript library](https://github.com/josdejong/jsonrepair) that automatically repairs malformed JSON documents. It handles JSON commonly found in LLM outputs, JavaScript code snippets, and various JSON-like formats.

**Module:** `github.com/kaptinlin/jsonrepair`
**Go Version:** 1.26
**Primary Use Case:** Fixing broken JSON from AI models, web scraping, and JavaScript sources

## Commands

```bash
# Run tests with race detection
task test

# Run linting (golangci-lint + go mod tidy check)
task lint

# Run both tests and linting
task all

# Clean build artifacts
task clean

# Run specific test
go test -race -run TestName

# Run tests with verbose output
go test -v -race ./...

# Run benchmarks
go test -bench=. -benchmem
```

## Architecture

### Recursive Descent Parser with Repair

The library uses a **single-pass recursive descent parser** that repairs JSON as it parses. The main entry point is `Repair()` (with deprecated alias `JSONRepair()`).

**Core parsing functions:**
- `parseValue()` - Dispatches to type-specific parsers
- `parseObject()` - Handles objects with automatic key/value repairs
- `parseArray()` - Handles arrays with automatic element repairs
- `parseString()` - Complex string parser handling quotes, escapes, concatenation, file paths
- `parseNumber()` - Number parsing with validation and leading-zero repairs
- `parseUnquotedString()` - Repairs unquoted strings, MongoDB calls, JSONP notation

### State Management

Parser operates on **runes ([]rune)** for proper Unicode handling. Two pointers track state:
- `i` (index) - Current position in input text (passed by pointer)
- `output` (strings.Builder) - Accumulated repaired JSON (passed by pointer)

### Repair Mechanisms

1. **Automatic insertion** - Missing commas, colons, quotes, brackets
2. **Automatic removal** - Trailing commas, invalid escapes, comments
3. **Character replacement** - Single quotes → double quotes, special whitespace → spaces
4. **Error recovery** - Truncated JSON completed by inserting missing closing tokens

### Special Parsing Modes

**String parsing with context-awareness** (jsonrepair.go:518-795):
- `stopAtDelimiter` - Stops at delimiters when end quote is missing
- `stopAtIndex` - Stops at specific index for precise repairs
- File path detection via `analyzePotentialFilePath()` - treats backslashes as literals in Windows paths

**Newline-delimited JSON** (jsonrepair.go:476-514):
- Detects NDJSON format when trailing comma/newline precedes a value
- Wraps multiple JSON values in array brackets

**Concatenated strings** (jsonrepair.go:797-846):
- Repairs JavaScript-style string concatenation: `"hello" + "world"` → `"helloworld"`

## Key Types and Interfaces

### Public API

```go
// Main repair function
func Repair(text string) (string, error)

// Deprecated alias
func JSONRepair(text string) (string, error)
```

### Error Handling

**Structured errors** (errors.go):
```go
type Error struct {
    Message  string
    Position int
    Err      error // underlying error for errors.Is/As
}
```

**Sentinel errors for errors.Is():**
- `ErrUnexpectedEnd` - Unexpected end of JSON string
- `ErrObjectKeyExpected` - Object key expected
- `ErrColonExpected` - Colon expected
- `ErrInvalidCharacter` - Invalid character
- `ErrUnexpectedCharacter` - Unexpected character
- `ErrInvalidUnicode` - Invalid unicode character

**Parse function pattern:**
```go
// Returns (success bool, error)
// error is non-nil only for non-repairable issues
func parseValue(text *[]rune, i *int, output *strings.Builder) (bool, error)
```

## Coding Rules

### Parser Implementation Patterns

**Whitespace handling:**
- `parseWhitespace()` - Preserves and normalizes whitespace
- `parseWhitespaceAndSkipComments()` - Combines whitespace parsing with comment removal
- Special whitespace characters (U+00A0, U+2009, etc.) replaced with regular spaces

**Position tracking:**
- Always track position (`i`) for error reporting
- Use `prevNonWhitespaceIndex()` to find previous significant characters
- Errors include exact position for debugging

**Output building:**
- Use `strings.Builder` for efficient string concatenation
- `insertBeforeLastWhitespace()` - Inserts characters before trailing whitespace
- `stripLastOccurrence()` - Removes specific characters from output

### Error Handling

- Return structured `*Error` with position information
- Use sentinel errors for common error types
- Parse functions return `(bool, error)` where bool indicates success
- Only return error for non-repairable issues (matches TypeScript reference)

### Performance

- Single-pass processing with minimal backtracking
- Rune-based parsing for full Unicode support
- Pre-compiled regex patterns in const.go
- No recursive calls that could cause stack overflow

## Testing

### Test Patterns

Use testify/assert and testify/require:

```go
// Test that valid JSON remains unchanged
assertRepairEqual(t, `{"valid": "json"}`)

// Test repair transformations
assertRepair(t, `{name: 'John'}`, `{"name": "John"}`)
```

### Test Coverage

- Valid JSON preservation
- All repair capabilities (quotes, commas, escapes, etc.)
- Error cases with position tracking
- Unicode handling
- Edge cases (truncated input, nested structures)

### Running Tests

```bash
# All tests with race detection
task test

# Specific test
go test -race -run TestParseObject

# Verbose output
go test -v -race ./...
```

## Dependencies

**Production:**
- `github.com/go-json-experiment/json` - Experimental JSON v2 (used for validation in tests)

**Testing:**
- `github.com/stretchr/testify` v1.11.1 - Assertions and test utilities

**Note:** Uses experimental `github.com/go-json-experiment/json` instead of stdlib `encoding/json/v2` until API is finalized.

## golangci-lint Configuration

**Version:** v2.9.0 (managed via `.golangci.version`)

**Key linters enabled:**
- Core: errcheck, govet, staticcheck, unused
- Error handling: err113, errorlint, nilerr
- Code quality: gocritic, revive, unconvert
- Security: gosec (disabled for test files)
- Performance: prealloc, copyloopvar

**Test file exclusions:**
- gosec - Security checks disabled in tests
- noctx - Context checks disabled in tests
- revive - Some style checks relaxed in tests

## Agent Skills

Package-local skills available in `.agents/skills/`:

- **agent-md-creating** - Generate CLAUDE.md for Go projects
- **code-simplifying** - Refine and simplify Go code for clarity
- **committing** - Create conventional commits for Go packages
- **dependency-selecting** - Select Go dependencies from vetted libraries
- **go-best-practices** - Google Go coding best practices and style guide
- **linting** - Set up and run golangci-lint v2
- **modernizing** - Go code modernization guide (Go 1.20-1.26)
- **ralphy-initializing** - Initialize Ralphy AI coding loop configuration
- **ralphy-todo-creating** - Create Ralphy TODO.yaml task files
- **readme-creating** - Generate README.md for Go libraries
- **releasing** - Guide release process for Go packages
- **testing** - Write Go tests following best practices

Use these skills when working on related tasks in this package.
