# Architecture - FQL Elixir Implementation

This document describes the architecture of the FQL Elixir implementation.

## Overview

The FQL Elixir implementation is a complete rewrite of the Go version, maintaining the same core concepts while leveraging Elixir's unique features.

## Module Structure

### Core Modules

#### `Fql.KeyVal`

The core data structures module that defines all query and data types.

**Key Types:**
- `query/0` - Union type for all query types
- `key_value/0` - Key-value pairs
- `key/0` - Keys (directory + tuple)
- `directory/0` - Directory paths
- `tuple/0` - Tuples of elements
- `value/0` - Values (primitives, tuples, variables)
- `variable/0` - Schema placeholders
- Primitive types: int, uint, bool, float, string, bytes, uuid, vstamp

**Design Patterns:**
- Uses tagged tuples for type disambiguation (e.g., `{:int, 42}` vs `{:uint, 42}`)
- Maps instead of structs for flexibility
- Elixir protocols for polymorphic operations (future enhancement)

#### `Fql.KeyVal.Class`

Classifies queries based on their structure and purpose.

**Query Classes:**
- `:constant` - Simple set operations
- `:vstamp_key` - Versionstamp in key
- `:vstamp_val` - Versionstamp in value
- `:clear` - Clear operations
- `:read_single` - Single key reads
- `:read_range` - Range reads
- `{:invalid, reason}` - Invalid queries

**Algorithm:**
- Traverses query structure recursively
- Accumulates attributes (variables, vstamps, clears, nils)
- Classifies based on attribute combinations

#### `Fql.KeyVal.Values`

Serialization and deserialization of values for FoundationDB storage.

**Features:**
- Big-endian and little-endian support
- Type-specific packing/unpacking
- Binary pattern matching for efficient encoding
- Error handling with descriptive messages

**Supported Operations:**
- `pack/3` - Serialize value to binary
- `unpack/3` - Deserialize binary to value

### Parser

#### `Fql.Parser.Scanner`

Tokenizes FQL query strings.

**Token Types:**
- Special characters: `/`, `=`, `(`, `)`, `<`, `>`, etc.
- Whitespace and newlines
- Escape sequences
- Other (identifiers, numbers, etc.)

**Implementation:**
- Recursive tokenization
- String-based scanning (no regex for core logic)
- Clear error messages

#### `Fql.Parser`

Converts token streams into query structures.

**Parser States:**
- Directory parsing
- Tuple parsing
- Value parsing
- Variable parsing
- String parsing

**Features:**
- Recursive descent parser
- State machine-based
- Supports all FQL syntax constructs

### Engine

#### `Fql.Engine`

Executes queries against FoundationDB.

**Core Functions:**
- `set/2` - Write operations
- `clear/2` - Clear operations
- `read_single/3` - Single key reads
- `read_range/3` - Range reads
- `list_directory/2` - Directory operations
- `transact/2` - Transaction wrapper

**Configuration:**
- Byte order (big/little endian)
- Logger
- Database connection

**Design Notes:**
- Uses erlfdb for FoundationDB access
- Maintains transaction context
- Supports both immediate and transactional execution

### CLI

#### `Fql.CLI`

Command-line interface for FQL.

**Modes:**
- Interactive mode - REPL-style interface
- Query mode - Execute queries from command line
- Help/Version - Information display

**Features:**
- Argument parsing with OptionParser
- Interactive loop with line input
- Query execution and result formatting
- Error handling and display

## Key Design Decisions

### 1. Tagged Tuples vs Structs

**Decision:** Use tagged tuples for primitive types

**Rationale:**
- Allows disambiguation between int/uint
- More idiomatic Elixir
- Pattern matching friendly
- Lower memory overhead

### 2. Maps vs Structs for Query Types

**Decision:** Use maps for key, key_value, etc.

**Rationale:**
- Flexibility for dynamic construction
- No need for defstruct boilerplate
- Easy to extend
- Pattern matching still works well

### 3. Error Handling

**Decision:** Use `{:ok, result}` / `{:error, reason}` tuples

**Rationale:**
- Idiomatic Elixir
- Forces error handling
- Clear error propagation
- Works well with `with` statements

### 4. Parser Implementation

**Decision:** Recursive descent parser with explicit state

**Rationale:**
- Clear and maintainable
- Easy to debug
- Matches Go implementation structure
- Good error messages

### 5. Engine Interface

**Decision:** Separate functions for each operation class

**Rationale:**
- Type safety at function level
- Clear API surface
- Easy to document
- Prevents misuse

## Differences from Go Version

### 1. No Visitor Pattern

**Go:** Uses generated visitor pattern for type operations

**Elixir:** Uses pattern matching and protocols

**Benefit:** More concise, more idiomatic

### 2. Concurrency Model

**Go:** Goroutines and channels

**Elixir:** Processes and message passing (future)

**Benefit:** Better fault tolerance, supervision trees

### 3. Type System

**Go:** Static typing with interfaces

**Elixir:** Dynamic typing with typespecs

**Tradeoff:** Less compile-time safety, more runtime flexibility

### 4. Transaction Management

**Go:** Context-based with defer

**Elixir:** Process-based with supervision

**Benefit:** Better error recovery, cleaner resource management

### 5. Code Generation

**Go:** Requires `go generate` for visitor pattern

**Elixir:** No code generation needed

**Benefit:** Simpler build process

## Future Enhancements

### 1. Protocol-Based Operations

Implement Elixir protocols for:
- Serialization
- Formatting
- Validation

### 2. GenServer for Engine

Use GenServer for:
- Connection pooling
- State management
- Concurrent query execution

### 3. Streaming Results

Use Elixir streams for:
- Large range queries
- Memory-efficient iteration
- Lazy evaluation

### 4. Macro-Based Query Construction

Domain-specific language using macros:
```elixir
import Fql.Query

query do
  dir("/path/to/key")
  tuple([1, 2, var(:int)])
  value(var(:string))
end
```

### 5. Pattern Matching in Queries

Enhanced pattern matching:
```elixir
case Fql.read(engine, query) do
  {:ok, %{value: {:int, n}}} when n > 0 -> ...
  {:ok, %{value: {:string, s}}} -> ...
end
```

## Testing Strategy

### Unit Tests

- KeyVal type construction
- Class classification
- Value serialization/deserialization
- Scanner tokenization
- Parser query construction

### Integration Tests (Future)

- FoundationDB interaction
- Transaction handling
- Error recovery
- Concurrent operations

### Property Tests (Future)

- Serialization round-tripping
- Query classification consistency
- Parser/formatter inverse
- Type safety properties

## Performance Considerations

### 1. Binary Pattern Matching

Elixir's binary pattern matching is highly optimized:
- Zero-copy operations where possible
- Compiled to efficient bytecode
- Fast type checking

### 2. Immutability

All data structures are immutable:
- Safe concurrent access
- Efficient structural sharing
- GC-friendly

### 3. Process Isolation

Future GenServer implementation:
- Isolated state per connection
- No shared memory contention
- Better error recovery

### 4. Lazy Evaluation

Streams for large result sets:
- Constant memory usage
- Composable operations
- Early termination

## Dependencies

### Runtime

- `erlfdb` - FoundationDB Erlang client
- `optimus` - CLI argument parsing

### Development

- `ex_unit` - Testing framework
- `dialyxir` - Type checking (future)
- `credo` - Code analysis (future)
- `ex_doc` - Documentation generation (future)

## Building and Deployment

### Development

```bash
cd elixir/fql
mix deps.get
mix compile
mix test
```

### Release

```bash
mix escript.build
```

Produces a self-contained executable.

### Docker (Future)

Containerized deployment with FoundationDB.

## Migration from Go

For users familiar with the Go version:

| Go | Elixir | Notes |
|---|---|---|
| `keyval.Int(42)` | `KeyVal.int(42)` | Returns `{:int, 42}` |
| `keyval.Variable{}` | `KeyVal.variable([])` | Returns map |
| `class.Classify()` | `Class.classify()` | Same logic |
| `values.Pack()` | `Values.pack()` | Returns tuple |
| `engine.Set()` | `Engine.set()` | Similar API |
| `parser.Parse()` | `Parser.parse()` | Returns tuple |

## Conclusion

This Elixir implementation maintains the core architecture and concepts of the Go version while leveraging Elixir's strengths in pattern matching, functional programming, and fault tolerance. The result is a more concise codebase that's easier to extend and maintain.
