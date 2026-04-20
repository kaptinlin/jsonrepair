# CLAUDE.md

This file provides operational guidance to Claude Code when working with the `jsonrepair` package.

## Project Overview

`jsonrepair` is a Go port of the JavaScript `jsonrepair` library. Package design, API, architecture, and coding rules now live in `SPECS/`.

## SPECS Index

- [SPECS/00-overview.md](SPECS/00-overview.md) — package scope, supported repair domain, and non-goals
- [SPECS/20-api-specs.md](SPECS/20-api-specs.md) — public API and error contract
- [SPECS/40-architecture-specs.md](SPECS/40-architecture-specs.md) — parser topology, state model, and repair modes
- [SPECS/50-coding-standards.md](SPECS/50-coding-standards.md) — implementation, testing, dependency, and tooling rules

Read the relevant spec before changing package behavior.

## Commands

```bash
# Run tests with race detection
task test

# Run linting (golangci-lint + go mod tidy check)
task lint

# Run both tests and linting
task

# Clean build artifacts
task clean

# Run specific test
go test -race -run TestName

# Run tests with verbose output
go test -v -race ./...

# Run benchmarks
go test -bench=. -benchmem
```

## Working Rules

- Keep design and internal package constraints in `SPECS/`, not `README.md`.
- Preserve the small public API unless a real user-facing need requires expansion.
- Run `task test` and `task lint` before finishing work.

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
