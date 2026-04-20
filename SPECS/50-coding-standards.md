# Coding Standards

## Overview

Implementation changes must preserve the parser's error-reporting accuracy, Unicode handling, and single-pass performance profile. Tests and tooling must reinforce those guarantees.

## Parser Implementation Rules

Whitespace handling must continue to normalize special whitespace through dedicated helpers such as `parseWhitespace` and `parseWhitespaceAndSkipComments`.

Position tracking must use the parser index and helper functions such as `prevNonWhitespaceIndex` so repair errors stay precise.

Output construction must use `strings.Builder` and helper functions that modify trailing output intentionally, such as `insertBeforeLastWhitespace` and `stripLastOccurrence`.

> **Why**: The parser is repair-oriented, so seemingly small helper changes can corrupt output or shift error positions. Centralizing these patterns keeps fixes predictable.
>
> **Rejected**: Ad hoc string concatenation, scattered whitespace normalization, and helper-free mutation of trailing output.

## Error Handling Rules

- Return structured `*Error` values with position information for non-repairable failures.
- Keep sentinel errors for common failure classes.
- Internal parse helpers should continue to use `(bool, error)` to distinguish repaired success from terminal failure.

## Performance Rules

- Preserve single-pass processing with minimal backtracking.
- Keep rune-based parsing for full Unicode support.
- Prefer precompiled regex patterns where regex-based checks are needed.
- Avoid recursion patterns that can grow into stack-overflow risks.

## Testing Rules

Tests must use `testify/assert` and `testify/require`.

Coverage must continue to emphasize:

- Preservation of already-valid JSON.
- Repair transformations across the supported defect classes.
- Position-aware error cases.
- Unicode behavior.
- Edge cases such as truncated input and nested structures.

> **Why**: The package earns trust by preserving good input while repairing bad input deterministically.
>
> **Rejected**: Narrow happy-path tests and documentation-layout tests that do not protect runtime behavior.

## Dependency and Tooling Rules

- `github.com/go-json-experiment/json` remains the production dependency for JSON-v2 behavior used in tests until the standard library API is finalized.
- `github.com/stretchr/testify` remains the default assertion library for tests.
- Linting must continue to run through `golangci-lint` with the version pinned in `.golangci.version`.
- Markdown under `SPECS/` must stay within the normal markdownlint coverage for the repository.

## Forbidden

- Do not return unstructured errors where position-aware errors are required.
- Do not replace rune-oriented parsing with byte-oriented parsing.
- Do not add tests whose only purpose is to pin `SPECS/` layout or `README`/`CLAUDE`/`AGENTS` file contents.
- Do not bypass `task lint` or `task test` before shipping changes.

## Acceptance Criteria

- [ ] Parser helpers follow the existing state, whitespace, and output-construction patterns.
- [ ] Tests protect repair behavior rather than documentation layout.
- [ ] Lint, tests, and markdownlint continue to cover the spec tree.

**Origin:** Migrated from `CLAUDE.md` (Coding Rules, Testing, Dependencies, and golangci-lint Configuration).
