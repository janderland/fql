# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FQL is a query language and alternative client API for FoundationDB written in Go. It provides textual description of key-value schemas and a Go API structurally equivalent to the query language. The project uses Go 1.19+ and FoundationDB client library v6.2.0+.

## Development Commands

### Building, Linting & Testing
- **Build, lint, and test**: `./build.sh --verify` (recommended - runs full CI/CD pipeline)
- **Generate code**: `./build.sh --generate` (checks `go generate ./...` output)
- **Build Docker images**: `./build.sh --image build,fql`
- **Run FQL interactively**: `./build.sh --run [args]`

### Docker-based Development
The build system uses fenv (https://github.com/janderland/fenv), a Docker-based FoundationDB development environment, for consistent builds and testing with FoundationDB integration. Always use `./build.sh --verify` instead of running Go commands directly to ensure proper environment setup.

## Architecture

### Core Components

**Engine (`/engine/`)**
- `engine.go`: Main query execution engine with methods for each query class
- `facade/`: Transaction and database abstraction layer
- `stream/`: Handles streaming results for range queries
- `internal/`: Internal query handling logic

**Parser (`/parser/`)**
- `parser.go`: Converts FQL query strings into keyval structures
- `scanner/`: Tokenizes FQL queries
- `internal/`: Internal parsing logic and AST building

**KeyVal (`/keyval/`)**
- Core data structures representing FQL queries and FoundationDB key-values
- `class/`: Query classification (Set, Clear, ReadSingle, ReadRange, Directory)
- `convert/`: Conversion between FQL and FoundationDB formats
- `tuple/`: Tuple comparison and manipulation
- `values/`: Value serialization and type handling

**Applications (`/internal/app/`)**
- `app.go`: Main CLI application entry point using Cobra
- `fullscreen/`: Interactive TUI mode using Bubble Tea
- `headless/`: Non-interactive command execution

### Key Architectural Patterns

1. **Query Classification**: All queries are classified into specific types (Set, Clear, ReadSingle, ReadRange, Directory) in the `keyval/class` package
2. **Facade Pattern**: Database operations abstracted through `engine/facade` for testing and transaction management
3. **Streaming**: Range queries use streaming pattern in `engine/stream` for efficient memory usage
4. **Parser State Machine**: FQL parsing uses explicit state machine in `parser/parser.go`
5. **Visitor Pattern**: Extensively used as a replacement for tagged unions (which Go lacks). Generated visitor interfaces (`TupleOperation`, `ValueOperation`, etc.) in `keyval/keyval_*.g.go` handle type-specific operations like serialization (`keyval/values`), formatting (`parser/format`), and tuple comparison (`keyval/tuple`). Each type implements acceptor methods that dispatch to appropriate visitor methods.

## FQL Query Language

FQL supports 5 query types:
- **Set**: Write key-value pairs
- **Clear**: Delete key-value pairs
- **ReadSingle**: Read single key-value
- **ReadRange**: Read multiple key-values with prefix matching
- **Directory**: List directory structures

Query syntax uses directories (`/path/to/dir`), tuples `("elem1", 2, 0xFF)`, and variables `<type|type>`.

## Testing Strategy

- Unit tests throughout codebase (files ending `_test.go`)
- Example tests demonstrate usage patterns
- Integration tests require FoundationDB connection
- Always use `./build.sh --verify` for full test suite with FDB container

## Code Generation

- Generated files: `keyval/*_query.g.go`, `keyval/keyval_*.g.go`
- Verify generation: `./build.sh --generate`
- Regenerate: `go generate ./...`

## Docker Environment

The build system uses fenv for consistent builds:
- `build.sh` script provides unified interface wrapping `fenv/fenv.sh`
- `docker/Dockerfile.builder` extends fenv base image with project-specific build tools (Go, golangci-lint, pandoc)
- `docker/Dockerfile` is a multi-stage build for the final fql runtime image:
  - `builder` stage extends fenv base image (same as Dockerfile.builder)
  - `gobuild` stage compiles the FQL binary
  - Final stage creates minimal runtime image
- `compose.yaml` defines runtime service for running fql image
- fenv automatically manages FDB container for integration testing