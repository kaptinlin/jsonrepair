# Post-Merge Nesting Depth Improvements

## Status

Implemented after merging [PR #19](https://github.com/kaptinlin/jsonrepair/pull/19) on August 23, 2026.

PR #19 establishes the initial denial-of-service protection by bounding recursive parsing. This follow-up keeps that protection while making the nesting boundary deterministic, propagating depth errors through NDJSON parsing, and documenting the public error contract.

This file is an implementation plan, not a canonical package specification. Durable behavior introduced by the follow-up must also be recorded in `SPECS/`.

## Goals

- Preserve protection against process-ending stack overflows.
- Define nesting depth in terms of open recursive structures rather than `parseValue` call count.
- Apply the same boundary to empty and non-empty structures.
- Return `ErrMaxDepthExceeded` from every parse path, including later NDJSON entries.
- Preserve the guarantee that a nil error always accompanies valid JSON output.
- Keep the existing public entry point and avoid configurable parser options.

## Non-Goals

- Replacing the recursive-descent parser with an iterative parser.
- Adding a public option for configuring the nesting limit.
- Changing unrelated repair behavior.
- Adding a second validation or repair pipeline.

## Depth Semantics

Use an internal constant named `maxNestingDepth` with the existing value of `10000`.

The `depth` argument means the number of recursive structures currently open before parsing the next value:

- A top-level value starts at depth `0`.
- Entering an object, array, or function-call wrapper consumes one level.
- Scalars do not consume another level.
- A structure may be opened only when `depth < maxNestingDepth`.
- The error position is the rune offset of the object, array, or function-call opener that would exceed the limit.
- Every NDJSON entry starts independently at depth `0`.

This definition allows a scalar or empty structure at the deepest permitted level and rejects the next structure consistently.

## Implementation Plan

### 1. Check Depth When Entering Structures

Remove the general depth check from the start of `parseValue`. Keep threading `depth` explicitly through internal helpers.

In `parseObject` and `parseArray`, check the current depth after recognizing the opening token and before writing output or advancing the input index. Return `newMaxDepthExceededError(*i)` when `depth >= maxNestingDepth`. Parse child values with `depth+1`.

In `parseUnquotedStringWithMode`, perform the same check after recognizing a function-call opening parenthesis. Report the position of the parenthesis and parse the wrapped value with `depth+1`.

Do not introduce a package-global or mutable shared depth counter. Explicit parameters keep parallel calls to `Repair` independent and avoid increment/decrement cleanup paths.

### 2. Propagate NDJSON Errors

Change `parseNewlineDelimitedJSON` to return an error.

- Return parse errors immediately instead of treating them as the end of the NDJSON sequence.
- Preserve the existing normal-stop behavior when no additional value is processed.
- Keep the depth reset to `0` for each entry.
- Make `Repair` return `"", err` when the helper fails.

Partial builder output does not need to be repaired after an error because `Repair` discards it.

### 3. Preserve the Structured Error Contract

Keep `ErrMaxDepthExceeded` and ensure failures return a position-aware `*Error` that supports both `errors.Is` and `errors.As`.

Update:

- `SPECS/20-api-specs.md` with the new sentinel error.
- `SPECS/40-architecture-specs.md` with the nesting-depth state definition.
- `SPECS/50-coding-standards.md` with the bounded-recursion requirement, if the existing rule is not sufficiently explicit.

## Test Plan

Add table-driven coverage for arrays, objects, and function calls:

- Exactly `maxNestingDepth` structures succeed.
- `maxNestingDepth+1` structures return `ErrMaxDepthExceeded`.
- Empty and scalar-leaf forms have the same boundary.
- Fully closed and truncated forms enforce the same limit.
- The returned `*Error.Position` points to the first disallowed opener.

Add NDJSON regression coverage:

- An over-depth first entry returns `ErrMaxDepthExceeded`.
- An over-depth second or later entry returns `ErrMaxDepthExceeded`.
- No over-depth NDJSON input returns invalid JSON with a nil error.

Extend the error-contract tests to cover `errors.Is`, `errors.As`, the sentinel stored in `Error.Err`, and the human-readable message.

Remove the in-process two-million-level test because disabling the guard would crash the entire test process. The exact-boundary tests already detect a missing guard safely. If an end-to-end crash-resistance test is retained, run the hazardous input in a subprocess so a regression becomes a normal parent-test failure.

## Suggested Commit Sequence

1. Add failing exact-boundary and NDJSON error-propagation tests.
2. Move depth enforcement to object, array, and function-call entry points.
3. Propagate errors from `parseNewlineDelimitedJSON` through `Repair`.
4. Update the error-contract tests and `SPECS/` documentation.
5. Run formatting, race tests, and linting.

## Acceptance Criteria

- [x] The process cannot reach unbounded parser recursion through objects, arrays, or function calls.
- [x] Empty and non-empty structures share the same maximum nesting boundary.
- [x] Every over-depth path returns a structured `ErrMaxDepthExceeded` error.
- [x] NDJSON parsing never converts a depth error into success or another error category.
- [x] Every successful `Repair` result in the new regression cases is valid JSON.
- [x] No new public configuration or parser entry point is introduced.
- [x] `task test` passes.
- [x] `task lint` passes.
