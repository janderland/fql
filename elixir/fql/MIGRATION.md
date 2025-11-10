# Migration Guide: Go to Elixir

This guide helps users migrate from the Go implementation of FQL to the Elixir version.

## Quick Start

### Installation

**Go version:**
```bash
go install github.com/janderland/fql
```

**Elixir version:**
```bash
cd elixir/fql
mix deps.get
mix escript.build
```

### Running Queries

Both versions support the same command-line interface:

```bash
# Interactive mode
./fql

# Non-interactive query
./fql -q "/my/dir(1, 2) = value"

# With flags
./fql -w -q "/my/dir(1, 2) = value"
```

## API Comparison

### Creating Queries

**Go:**
```go
import "github.com/janderland/fql/keyval"

kv := keyval.KeyValue{
    Key: keyval.Key{
        Directory: keyval.Directory{"path", "to", "dir"},
        Tuple: keyval.Tuple{
            keyval.Int(1),
            keyval.Int(2),
        },
    },
    Value: keyval.String("value"),
}
```

**Elixir:**
```elixir
alias Fql.KeyVal

kv = KeyVal.key_value(
  KeyVal.key(
    ["path", "to", "dir"],
    [KeyVal.int(1), KeyVal.int(2)]
  ),
  "value"
)
```

### Creating Variables

**Go:**
```go
variable := keyval.Variable{keyval.IntType, keyval.StringType}
```

**Elixir:**
```elixir
variable = KeyVal.variable([:int, :string])
```

### Using Primitive Types

**Go:**
```go
intVal := keyval.Int(42)
uintVal := keyval.Uint(100)
strVal := keyval.String("hello")
bytesVal := keyval.Bytes([]byte{1, 2, 3})
uuidVal := keyval.UUID{...}
```

**Elixir:**
```elixir
int_val = KeyVal.int(42)          # {:int, 42}
uint_val = KeyVal.uint(100)       # {:uint, 100}
str_val = "hello"                  # strings are strings
bytes_val = KeyVal.bytes(<<1, 2, 3>>)  # {:bytes, <<1, 2, 3>>}
uuid_val = KeyVal.uuid(<<...::128>>)    # {:uuid, <<...::128>>}
```

**Key Difference:** Elixir uses tagged tuples for disambiguation

### Parsing Queries

**Go:**
```go
import "github.com/janderland/fql/parser"

p := parser.New()
query, err := p.Parse("/my/dir(1, 2) = value")
if err != nil {
    // handle error
}
```

**Elixir:**
```elixir
alias Fql.Parser

case Parser.parse("/my/dir(1, 2) = value") do
  {:ok, query} -> # use query
  {:error, reason} -> # handle error
end
```

### Classifying Queries

**Go:**
```go
import "github.com/janderland/fql/keyval/class"

queryClass := class.Classify(kv)
switch queryClass {
case class.Constant:
    // handle constant
case class.ReadSingle:
    // handle read single
// ...
}
```

**Elixir:**
```elixir
alias Fql.KeyVal.Class

case Class.classify(kv) do
  :constant -> # handle constant
  :read_single -> # handle read single
  {:invalid, reason} -> # handle invalid
end
```

### Executing Queries

**Go:**
```go
import (
    "github.com/janderland/fql/engine"
    "github.com/janderland/fql/engine/facade"
)

eg := engine.New(
    facade.NewTransactor(db, directory.Root()),
    engine.ByteOrder(binary.BigEndian),
)

err := eg.Set(kv)
```

**Elixir:**
```elixir
alias Fql.Engine

engine = Engine.new(db, byte_order: :big)

case Engine.set(engine, kv) do
  :ok -> # success
  {:error, reason} -> # handle error
end
```

### Reading Data

**Go:**
```go
// Single read
result, err := eg.ReadSingle(query, engine.SingleOpts{
    Filter: true,
})

// Range read
results, err := eg.ReadRange(query, engine.RangeOpts{
    Reverse: false,
    Filter: true,
    Limit: 100,
})
```

**Elixir:**
```elixir
# Single read
case Engine.read_single(engine, query, %{filter: true}) do
  {:ok, result} -> # use result
  {:error, reason} -> # handle error
end

# Range read
case Engine.read_range(engine, query, %{
  reverse: false,
  filter: true,
  limit: 100
}) do
  {:ok, results} -> # use results
  {:error, reason} -> # handle error
end
```

### Transactions

**Go:**
```go
result, err := eg.Transact(func(txEg engine.Engine) (interface{}, error) {
    if err := txEg.Set(kv1); err != nil {
        return nil, err
    }
    if err := txEg.Set(kv2); err != nil {
        return nil, err
    }
    return nil, nil
})
```

**Elixir:**
```elixir
Engine.transact(engine, fn tx_engine ->
  with :ok <- Engine.set(tx_engine, kv1),
       :ok <- Engine.set(tx_engine, kv2) do
    {:ok, nil}
  end
end)
```

## Type Mapping

| Go Type | Elixir Type | Notes |
|---------|-------------|-------|
| `keyval.Int` | `{:int, integer()}` | Tagged tuple |
| `keyval.Uint` | `{:uint, non_neg_integer()}` | Tagged tuple |
| `keyval.Bool` | `boolean()` | Native Elixir |
| `keyval.Float` | `float()` | Native Elixir |
| `keyval.String` | `String.t()` | Native Elixir |
| `keyval.Bytes` | `{:bytes, binary()}` | Tagged tuple |
| `keyval.UUID` | `{:uuid, <<_::128>>}` | Tagged tuple |
| `keyval.Nil` | `:nil` | Atom |
| `keyval.Variable` | `%{types: [atom()]}` | Map |
| `keyval.MaybeMore` | `:maybe_more` | Atom |
| `keyval.Clear` | `:clear` | Atom |
| `keyval.VStamp` | `%{tx_version: ..., user_version: ...}` | Map |
| `keyval.VStampFuture` | `%{user_version: ...}` | Map |

## Error Handling

**Go:**
```go
result, err := someOperation()
if err != nil {
    return err
}
// use result
```

**Elixir:**
```elixir
# Pattern matching
case some_operation() do
  {:ok, result} -> # use result
  {:error, reason} -> # handle error
end

# With statement (for chaining)
with {:ok, result1} <- operation1(),
     {:ok, result2} <- operation2(result1) do
  {:ok, result2}
end
```

## Query Language

The query language syntax is **identical** between Go and Elixir versions:

```
# Set a value
/path/to/dir(1, 2, 3) = "value"

# Read a single key
/path/to/dir(1, 2, 3)

# Range read with variable
/path/to/dir(<int>, 2, 3)

# Clear a key
/path/to/dir(1, 2, 3) = clear

# Variable with type constraints
/path/to/dir(<int|uint>) = <string>

# List directories
/path/to/dir
```

## Command-Line Flags

All flags are the same between versions:

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help |
| `--version` | `-v` | Show version |
| `--query` | `-q` | Execute query |
| `--cluster` | `-c` | Cluster file path |
| `--write` | `-w` | Allow writes |
| `--little` | `-l` | Little endian |
| `--reverse` | `-r` | Reverse order |
| `--strict` | `-s` | Strict mode |
| `--bytes` | `-b` | Show full bytes |
| `--log` | | Enable logging |
| `--log-file` | | Log file path |
| `--limit` | | Result limit |

## Performance Considerations

### Memory Usage

**Go:**
- Lower baseline memory
- Manual memory management
- Struct overhead

**Elixir:**
- Higher baseline (BEAM VM)
- Garbage collected
- Immutable data structures
- Structural sharing

### Concurrency

**Go:**
- Goroutines (lightweight threads)
- Shared memory with mutexes
- Channel-based communication

**Elixir:**
- Processes (even lighter)
- No shared memory
- Message passing
- Better fault isolation

### Startup Time

**Go:** Faster startup (~milliseconds)

**Elixir:** Slower startup (~100ms BEAM init)

## Migration Strategy

### Phase 1: Parallel Running

Run both versions side-by-side:

```bash
# Test with Go version
fql-go -q "/test(1) = value"

# Verify with Elixir version
fql-elixir -q "/test(1)"
```

### Phase 2: Gradual Migration

1. Start with read-only operations
2. Migrate write operations
3. Test thoroughly
4. Switch production traffic

### Phase 3: Decommission

Remove Go version after validation period.

## Common Pitfalls

### 1. Type Tagging

**Problem:** Forgetting to tag int/uint

**Go:**
```go
keyval.Int(42)  // Correct
```

**Elixir:**
```elixir
42                  # Wrong - untagged
KeyVal.int(42)      # Correct - returns {:int, 42}
```

### 2. Error Handling

**Problem:** Not pattern matching on results

**Wrong:**
```elixir
result = Parser.parse(query)
# result might be {:error, ...}
```

**Correct:**
```elixir
case Parser.parse(query) do
  {:ok, result} -> use result
  {:error, _} -> handle error
end
```

### 3. Binary Strings

**Problem:** String vs binary confusion

**Elixir:**
```elixir
"hello"              # String (UTF-8)
<<"hello">>          # Also a string
<<1, 2, 3>>          # Binary
KeyVal.bytes(<<...>>) # Tagged bytes
```

### 4. Nil Values

**Problem:** Using Go's nil

**Go:**
```go
var x *keyval.KeyValue = nil  // pointer nil
keyval.Nil{}                   // FQL nil value
```

**Elixir:**
```elixir
nil                # Elixir nil (atom)
:nil               # FQL nil value
```

## Testing

### Unit Tests

**Go:**
```go
func TestParse(t *testing.T) {
    p := parser.New()
    result, err := p.Parse("/test")
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

**Elixir:**
```elixir
test "parses directory query" do
  assert {:ok, result} = Parser.parse("/test")
  assert is_list(result)
end
```

### Integration Tests

Both versions can use the same FoundationDB test cluster:

```bash
# Start FDB test cluster
docker run -d foundationdb/foundationdb:6.2.0

# Run tests
mix test              # Elixir
go test ./...         # Go
```

## Support and Resources

### Documentation

- **Go:** https://pkg.go.dev/github.com/janderland/fql
- **Elixir:** `mix docs` (generates local docs)

### Source Code

- **Go:** `/` (root directory)
- **Elixir:** `/elixir/fql/`

### Getting Help

For migration questions:
1. Check this guide
2. Review ARCHITECTURE.md
3. Compare equivalent code in both versions
4. Open an issue on GitHub

## Conclusion

The Elixir version maintains API compatibility at the query language level while providing a more functional, fault-tolerant implementation. Most code can be migrated by following the patterns in this guide.
