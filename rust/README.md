# FQL Rust Implementation

This directory contains a Rust rewrite of the FQL (FoundationDB Query Language) project. The Rust implementation leverages Rust's powerful type system, enums, and pattern matching to provide a more idiomatic and type-safe query language implementation.

## Project Structure

The Rust implementation is organized as a Cargo workspace with the following crates:

```
rust/
├── Cargo.toml          # Workspace configuration
├── keyval/             # Core data structures and types
├── parser/             # Query string parsing and formatting
├── engine/             # Query execution engine
└── cli/                # Command-line interface
```

### Crates

#### `keyval` - Core Data Structures

The keyval crate contains the fundamental types for FQL queries and key-values:

- **Query types**: `Query`, `KeyValue`, `Key`, `Directory`
- **Element types**: `TupElement`, `DirElement`, `Value`
- **Type system**: `Variable`, `ValueType`, `VStamp`, `VStampFuture`
- **Submodules**:
  - `class`: Query classification (Constant, Clear, ReadSingle, ReadRange, etc.)
  - `convert`: Type conversions between FQL and FoundationDB formats
  - `tuple`: Tuple comparison and matching logic
  - `values`: Value serialization and deserialization

**Key Improvement over Go**: Unlike the Go implementation which uses the visitor pattern to work around the lack of tagged unions, the Rust implementation uses native enums with pattern matching, resulting in cleaner, more maintainable code.

#### `parser` - Query Parsing

The parser crate handles tokenization and parsing of FQL query strings:

- **scanner**: Tokenizes FQL queries into tokens
- **format**: Converts keyval structures back to FQL strings
- **Parser API**: `parse(input: &str) -> Result<Query, ParseError>`

**Example**:
```rust
use parser::parse;

let query = parse("/users/(42)=<>")?;
```

#### `engine` - Query Execution

The engine crate executes FQL queries against FoundationDB:

- **Engine**: Main query executor with async/await support
- **facade**: Database abstraction layer for testing
- **stream**: Async streaming support for range queries

**Key Features**:
- Async/await for all database operations
- Trait-based database abstraction
- Streaming results for range queries using Rust's `Stream` trait

**Example**:
```rust
use engine::{Engine, EngineConfig};

let engine = Engine::new(EngineConfig::default());
let results = engine.execute(&query).await?;
```

#### `cli` - Command-Line Interface

The CLI crate provides both interactive and non-interactive modes:

- Interactive TUI mode (to be implemented with ratatui)
- Single query execution mode
- Based on clap for argument parsing

**Example**:
```bash
# Execute a single query
cargo run --bin cli -- "/users/(42)=<>"

# Interactive mode
cargo run --bin cli -- interactive
```

## Key Advantages of the Rust Implementation

### 1. **Type Safety with Enums**

The Go implementation uses interfaces and the visitor pattern to simulate tagged unions. Rust's native enum support makes this much cleaner:

**Go (visitor pattern)**:
```go
type Value interface {
    Value(ValueOperation)
}

type ValueOperation interface {
    ForInt(Int)
    ForString(String)
    // ... many more methods
}
```

**Rust (native enums)**:
```rust
pub enum Value {
    Int(i64),
    String(String),
    // ... other variants
}

match value {
    Value::Int(i) => handle_int(i),
    Value::String(s) => handle_string(s),
    // Compiler ensures all cases are handled
}
```

### 2. **Async/Await**

The Rust implementation uses native async/await for all database operations, providing:
- Better ergonomics than Go's goroutines for this use case
- Structured concurrency
- Zero-cost abstractions

### 3. **Ownership and Borrowing**

Rust's ownership system prevents entire classes of bugs:
- No null pointer exceptions
- No data races
- Memory safety without garbage collection

### 4. **Pattern Matching**

Exhaustive pattern matching ensures all cases are handled at compile time:
```rust
match query {
    Query::KeyValue(kv) => handle_keyvalue(kv),
    Query::Key(key) => handle_key(key),
    Query::Directory(dir) => handle_directory(dir),
    // Compiler error if any variant is missing
}
```

### 5. **Trait System**

The facade module uses traits for database abstraction, similar to Go interfaces but with:
- Static dispatch (zero runtime cost)
- Associated types
- Const generics

## Building and Testing

### Prerequisites

- Rust 1.70+ (2021 edition)
- FoundationDB 6.2+ (for full integration)

### Build

```bash
# Build all crates
cargo build

# Build with optimizations
cargo build --release

# Build specific crate
cargo build -p keyval
```

### Test

```bash
# Run all tests
cargo test

# Run tests for specific crate
cargo test -p keyval

# Run tests with output
cargo test -- --nocapture
```

### Documentation

```bash
# Generate documentation
cargo doc --no-deps --open

# Generate documentation for all workspace crates
cargo doc --workspace --no-deps --open
```

## Differences from Go Implementation

### Architecture

| Aspect | Go Implementation | Rust Implementation |
|--------|------------------|-------------------|
| Polymorphism | Interfaces + Visitor Pattern | Enums + Pattern Matching |
| Concurrency | Goroutines + Channels | Async/Await + Futures |
| Error Handling | Multiple returns | Result<T, E> type |
| Memory Management | Garbage Collection | Ownership + Borrowing |
| Dependencies | Go modules | Cargo |

### Code Generation

The Go implementation uses code generation (`go generate`) to create visitor pattern boilerplate. The Rust implementation eliminates this entirely by using native enums.

**Go**: ~4 generated files in keyval/ (~500 LOC)
**Rust**: 0 generated files (enums are first-class)

### Lines of Code

Estimated comparison for core functionality:

| Component | Go LOC | Rust LOC | Reduction |
|-----------|--------|----------|-----------|
| keyval core | ~500 | ~300 | 40% |
| classification | ~250 | ~200 | 20% |
| parser/scanner | ~680 | ~400 | 41% |
| **Total** | ~1430 | ~900 | ~37% |

The Rust implementation is more concise while being more type-safe.

## Implementation Status

### ✅ Completed

- [x] Workspace structure and Cargo configuration
- [x] Core keyval data structures with enums
- [x] Query classification module
- [x] Scanner/tokenizer foundation
- [x] Format module for query formatting
- [x] Engine structure with async/await
- [x] Facade trait definitions
- [x] CLI application structure

### 🚧 In Progress / TODO

- [ ] Complete parser state machine implementation
- [ ] FoundationDB integration (requires fdb-rs crate)
- [ ] Tuple comparison logic implementation
- [ ] Value serialization/deserialization
- [ ] Type conversion utilities
- [ ] Streaming implementation for range queries
- [ ] Interactive TUI with ratatui
- [ ] Comprehensive test suite
- [ ] Integration tests with FoundationDB
- [ ] Performance benchmarks

## Dependencies

### Core Dependencies

- **serde**: Serialization/deserialization
- **uuid**: UUID support
- **thiserror**: Error handling
- **async-trait**: Async trait methods
- **futures**: Async primitives
- **tokio**: Async runtime

### CLI Dependencies

- **clap**: Command-line argument parsing
- **ratatui**: Terminal UI framework (for interactive mode)
- **crossterm**: Terminal manipulation

## Performance Considerations

The Rust implementation is expected to have several performance advantages:

1. **Zero-cost abstractions**: Rust's enums and pattern matching compile to efficient machine code
2. **No GC pauses**: Deterministic performance without garbage collection
3. **Inline expansion**: Extensive use of generics and monomorphization
4. **SIMD**: Potential for auto-vectorization in serialization code

Benchmarks to be added once implementation is complete.

## Contributing

The Rust implementation follows standard Rust idioms and best practices:

- Format code with `rustfmt`: `cargo fmt`
- Lint with Clippy: `cargo clippy`
- Document public APIs with doc comments (`///`)
- Write tests alongside code
- Use `Result<T, E>` for fallible operations
- Prefer borrowing over cloning where possible

## License

Same as the parent FQL project (MIT OR Apache-2.0).

## Acknowledgments

This Rust implementation is a rewrite of the original Go implementation at [github.com/janderland/fql](https://github.com/janderland/fql). The architecture and API design are based on the Go version, adapted to leverage Rust's unique features.
