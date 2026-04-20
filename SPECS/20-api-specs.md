# API Specifications

## Overview

The public API must stay small: callers hand the package a string and receive either repaired JSON or a structured error that explains why repair could not finish.

## Public Entry Point

```go
func Repair(text string) (string, error)
```

`Repair` must remain the primary public entry point. It accepts raw text and returns a repaired JSON document string or an error.

> **Why**: A single entry point keeps repair policy centralized and makes the library easy to adopt in one-shot cleanup flows.
>
> **Rejected**: Option-heavy constructors, streaming APIs, and exported parser internals. Those add configuration and compatibility burden without improving the dominant use case.

## Error Contract

Non-repairable failures must return a structured `*Error` with position information:

```go
type Error struct {
    Message  string
    Position int
    Err      error
}
```

The error surface must continue to support `errors.Is` through sentinel errors for the core failure classes:

- `ErrUnexpectedEnd`
- `ErrObjectKeyExpected`
- `ErrColonExpected`
- `ErrInvalidCharacter`
- `ErrUnexpectedCharacter`
- `ErrInvalidUnicode`

> **Why**: Callers need both a human-readable message and a machine-checkable category when repair cannot proceed.
>
> **Rejected**: Returning plain string errors only, or exposing parser-specific internal state as part of the public contract.

## Compatibility Boundary

- The package contract is the repaired JSON string and the structured error behavior.
- Internal helper names and parser decomposition are not part of the public compatibility surface.
- Error categories should remain stable enough for callers using `errors.Is`.

## Forbidden

- Do not export additional parser stages unless they are required by a real public use case.
- Do not return bare internal errors when a position-aware `*Error` is required.
- Do not make callers choose between separate validation and repair entry points for the common flow.

## Acceptance Criteria

- [ ] The package exposes a single primary repair function.
- [ ] Non-repairable failures can be inspected with `errors.As` and `errors.Is`.
- [ ] Public compatibility is defined in terms of behavior, not internal helper structure.

**Origin:** Migrated from `CLAUDE.md` (Key Types and Interfaces).
