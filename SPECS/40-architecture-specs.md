# Architecture Specifications

## Overview

The package must use a single-pass recursive-descent parser that repairs JSON while parsing. Repair is part of the parse path, not a preprocessing or postprocessing phase.

> **Why**: Inline repair keeps position tracking coherent, avoids duplicate parsing pipelines, and keeps the implementation fast enough for large AI-generated payloads.
>
> **Rejected**: A tokenizer-plus-rewriter pipeline and a separate repair pass after parsing. Those duplicate state and make recovery logic harder to reason about.

## Parser Topology

The parser must dispatch from `Repair` through type-specific parsing helpers.

Core helpers include:

- `parseValue`
- `parseObject`
- `parseArray`
- `parseString`
- `parseNumber`
- `parseUnquotedString`

Each helper is responsible for both recognizing its input form and applying the repairs that belong to that form.

## State Model

Parser state must operate on `[]rune` input so Unicode handling remains correct. The main mutable state is:

- `i` for the current input position.
- `output` as a `strings.Builder` holding the repaired JSON.

Position tracking is part of the architecture, not an optional debugging aid.

## Repair Strategy

Repairs must stay within four categories:

1. Automatic insertion of missing structural tokens.
2. Automatic removal of clearly invalid or extraneous syntax.
3. Character replacement where there is an unambiguous JSON-normalized form.
4. Error recovery for truncated input by closing incomplete structures.

## Specialized Modes

The architecture must preserve the targeted recovery modes that distinguish this package from a strict JSON parser:

- Context-aware string parsing that can stop at delimiters or a specific index and recognize literal Windows paths.
- Newline-delimited JSON detection that wraps multiple top-level values in an array.
- Concatenated-string repair for JavaScript-style `"a" + "b"` inputs.

## Forbidden

- Do not introduce a second repair pipeline before or after the parser.
- Do not switch the parser core from rune-based indexing to byte-based indexing.
- Do not treat specialized recovery modes as ad hoc output rewrites detached from parser state.

## Acceptance Criteria

- [ ] Repair decisions happen inside the parser flow.
- [ ] Unicode-safe position tracking remains part of the core state model.
- [ ] Specialized modes remain explicit architectural features rather than incidental patches.

**Origin:** Migrated from `CLAUDE.md` (Architecture).
