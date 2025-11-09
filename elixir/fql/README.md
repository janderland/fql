# FQL - Elixir Implementation

FQL is a query language and alternative client API for FoundationDB, rewritten in Elixir from the original Go implementation.

## Overview

This is a complete rewrite of the FQL project in Elixir, providing:

- A query language for FoundationDB with textual description of key-value schemas
- An Elixir API structurally equivalent to the query language
- Support for all FQL query types: Set, Clear, ReadSingle, ReadRange, and Directory queries

## Installation

### Prerequisites

- Elixir 1.14 or later
- Erlang/OTP 24 or later
- FoundationDB 6.2.0 or later

### Building

```bash
cd elixir/fql
mix deps.get
mix compile
```

### Building the CLI

```bash
mix escript.build
```

This will create an executable `fql` binary.

## Usage

### Interactive Mode

```bash
./fql
```

### Non-Interactive Mode

```bash
./fql --query "/path/to/key = value"
```

### Flags

- `-c, --cluster` - Path to FoundationDB cluster file
- `-q, --query` - Execute query non-interactively
- `-w, --write` - Allow write queries
- `-l, --little` - Use little endian encoding instead of big endian
- `-r, --reverse` - Query range-reads in reverse order
- `--limit` - Limit the number of KVs read in range-reads
- `-s, --strict` - Throw an error if a KV doesn't match the schema
- `-b, --bytes` - Print full byte strings instead of just their length

## Architecture

The Elixir implementation follows the same architectural patterns as the Go version:

### Modules

- **Fql.KeyVal** - Core data structures representing FQL queries and FoundationDB key-values
- **Fql.Parser** - Converts FQL query strings into KeyVal structures
- **Fql.Engine** - Query execution engine
- **Fql.CLI** - Command-line interface

### Key Differences from Go Version

- Uses Elixir protocols instead of Go's visitor pattern
- Leverages pattern matching for query classification
- Uses GenServer for transaction management
- Native support for tagged unions via Elixir's tagged tuples

## Testing

```bash
mix test
```

## Documentation

Generate documentation:

```bash
mix docs
```

## License

See LICENSE file in the root directory.
