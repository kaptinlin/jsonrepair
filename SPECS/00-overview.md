# Overview

## Scope

`jsonrepair` repairs malformed JSON and JSON-adjacent text into a valid JSON document string. The library is optimized for inputs commonly produced by LLMs, copied from JavaScript sources, or recovered from scraped content.

> **Why**: The package exists to recover useful data from near-JSON inputs without forcing every caller to maintain a custom cleanup pipeline.
>
> **Rejected**: General-purpose JavaScript parsing, schema-aware repair, and pluggable repair stages. Those expand the surface area beyond a small, predictable repair library.

## Supported Repair Domain

The package must handle the JSON-adjacent defects that recur in its target sources:

- Missing structural tokens such as commas, colons, quotes, and closing brackets.
- Quote and whitespace normalization, including single quotes, smart quotes, and special whitespace.
- Non-JSON wrappers and extensions such as comments, fenced code blocks, JSONP notation, MongoDB wrappers, concatenated strings, and newline-delimited JSON.
- Language literals that have an obvious JSON equivalent, such as Python-style `None`, `True`, and `False`.

## Output Guarantees

- The output must be valid JSON text representing exactly one JSON value.
- Already-valid JSON must continue to repair successfully without requiring callers to branch between validation and repair.
- When the input is newline-delimited JSON, the repaired result must be a JSON array containing the parsed values.

## Non-Goals

- The package is not a JavaScript interpreter.
- The package is not a schema validator.
- The package is not a formatting or pretty-printing tool.

## Forbidden

- Do not add repair rules that depend on executing embedded code.
- Do not widen the package into a configurable transformation framework.
- Do not move design rules back into `README.md`; `SPECS/` is the canonical home for package design constraints.

## Acceptance Criteria

- [ ] The package scope is limited to repairing malformed JSON and JSON-adjacent text.
- [ ] The supported repair domain is explicit enough to guide future feature decisions.
- [ ] Non-goals make it clear which adjacent requests belong outside this package.

**Origin:** Migrated from `CLAUDE.md` (Project Overview and repair-capability guidance).
