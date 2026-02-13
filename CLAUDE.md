# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **Go port of the jsonrepair JavaScript library**, designed to automatically fix malformed JSON documents. The library handles JSON content that commonly appears in LLM outputs, JavaScript code snippets, and various JSON-like formats.

## Common Development Commands

```bash
# Run all tests with race detection
make test

# Run linting (golangci-lint + go mod tidy check)
make lint

# Run both tests and linting
make all

# Clean build artifacts
make clean

# Run a single test
go test -race -run TestName

# Run tests with verbose output
go test -v -race ./...
```

## Architecture Overview

### Core Parsing Engine

The library uses a **recursive descent parser with repair capabilities**. The main entry point `Repair()` (with `JSONRepair()` as a deprecated alias) orchestrates parsing through specialized functions:

- **parseValue()** - Dispatches to specific type parsers (object, array, string, number, keywords)
- **parseObject()** - Handles object parsing with automatic key/value repairs
- **parseArray()** - Handles array parsing with automatic element repairs
- **parseString()** - Complex string parser handling quotes, escapes, concatenation, and file paths
- **parseNumber()** - Number parsing with validation and leading-zero repairs
- **parseUnquotedString()** - Repairs unquoted strings, MongoDB calls, and JSONP notation

### Repair Strategy

The parser operates on **runes ([]rune)** instead of bytes to properly handle Unicode. Two pointers track state:

- **i (index)** - Current position in input text
- **output (strings.Builder)** - Accumulated repaired JSON

Key repair mechanisms:

1. **Automatic insertion** - Missing commas, colons, quotes, brackets
2. **Automatic removal** - Trailing commas, invalid escape characters, comments
3. **Character replacement** - Single quotes to double quotes, special whitespace to spaces
4. **Error recovery** - Handles truncated JSON by inserting missing closing tokens

### Error Handling

The library uses **structured errors** (errors.go:18-76):

- **Error** struct with Message, Position, and wrapped error
- Predefined sentinel errors (ErrUnexpectedEnd, ErrInvalidUnicode, etc.)
- Supports errors.Is() / errors.As() for error checking

Parse functions return `(bool, error)` where:
- `bool` indicates if parsing succeeded
- `error` is non-nil only for **non-repairable issues** (matches TypeScript reference implementation)

### Special Parsing Modes

**String parsing with context-awareness** (jsonrepair.go:518-795):
- `stopAtDelimiter` - Stops at delimiters when end quote is missing
- `stopAtIndex` - Stops at a specific index for precise repairs
- File path detection via `analyzePotentialFilePath()` - treats backslashes as literal characters in Windows paths

**Newline-delimited JSON** (jsonrepair.go:476-514):
- Detects NDJSON format when trailing comma/newline precedes a value
- Wraps multiple JSON values in array brackets

**Concatenated strings** (jsonrepair.go:797-846):
- Repairs JavaScript-style string concatenation: `"hello" + "world"` → `"helloworld"`

## Testing Architecture

Tests use **testify/assert** and **testify/require** for assertions. Two main test patterns:

1. **assertRepairEqual(t, json)** - Tests that valid JSON remains unchanged
2. **assertRepair(t, input, expected)** - Tests repair transformations

Test coverage includes:
- Valid JSON preservation
- All repair capabilities (quotes, commas, escapes, etc.)
- Error cases with position tracking
- Unicode handling
- Edge cases (truncated input, nested structures)

## Code Quality Standards

### golangci-lint Configuration

This package uses golangci-lint v2.9.0 (managed via `.golangci.version`). The Makefile automatically installs the correct version in `./bin/`.

Enabled linters include:
- Core: errcheck, govet, staticcheck, unused
- Error handling: err113, errorlint, nilerr
- Code quality: gocritic, revive, unconvert
- Security: gosec (disabled for test files)
- Performance: prealloc, copyloopvar

### Development Patterns

**Whitespace handling**:
- `parseWhitespace()` - Preserves and normalizes whitespace
- `parseWhitespaceAndSkipComments()` - Combines whitespace parsing with comment removal
- Special whitespace characters (U+00A0, U+2009, etc.) replaced with regular spaces

**Position tracking**:
- Always track position (`i`) for error reporting
- Use `prevNonWhitespaceIndex()` to find previous significant characters
- Errors include exact position for debugging

**Output building**:
- Use `strings.Builder` for efficient string concatenation
- `insertBeforeLastWhitespace()` - Inserts characters before trailing whitespace
- `stripLastOccurrence()` - Removes specific characters from output

## Go Version & Dependencies

- **Go 1.26** - Uses modern Go features
- **github.com/stretchr/testify v1.11.1** - Testing framework

## Performance Characteristics

The parser is designed for **single-pass processing** with backtracking only when necessary for repairs. Key performance considerations:

- Rune-based parsing supports full Unicode
- String builder minimizes allocations
- Regex usage limited to pre-compiled patterns (const.go)
- No recursive calls that could cause stack overflow (fixed in commit 1a1037c)
