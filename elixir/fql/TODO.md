# TODO: Elixir Implementation Improvements

This document outlines improvements to make the FQL Elixir implementation more idiomatic, maintainable, and production-ready.

## High Priority - Core Idioms

### 1. Replace Maps with Structs
**Current:** Using plain maps for KeyValue, Key, Variable, VStamp, etc.
**Improvement:** Use `defstruct` for better pattern matching and compile-time guarantees

```elixir
# Current
%{key: key, value: value}

# Proposed
defmodule Fql.KeyVal.KeyValue do
  @type t :: %__MODULE__{
    key: Fql.KeyVal.Key.t(),
    value: Fql.KeyVal.value()
  }

  defstruct [:key, :value]
end

# Usage
%KeyValue{key: key, value: value}
```

**Benefits:**
- Compile-time field checking
- Better documentation
- Pattern matching improvements
- IDE autocomplete support

**Files to change:**
- `lib/fql/keyval.ex` - Define structs for KeyValue, Key, Variable, VStamp, VStampFuture
- All files using these types
- Update tests

**Estimated effort:** Medium (affects many files)

---

### 2. Implement Elixir Protocols
**Current:** Using tagged tuples and case statements for polymorphism
**Improvement:** Define protocols for common operations

```elixir
defprotocol Fql.Serializable do
  @doc "Serialize a value to bytes"
  def to_bytes(value, opts)
end

defimpl Fql.Serializable, for: Integer do
  def to_bytes(value, opts) do
    # Implement int serialization
  end
end

# Usage
Fql.Serializable.to_bytes(value, byte_order: :big)
```

**Protocols to define:**
- `Fql.Serializable` - For value serialization
- `Fql.Formattable` - For query formatting
- `Fql.Classifiable` - For query classification

**Benefits:**
- More extensible
- Idiomatic Elixir
- Open for extension without modifying core code

**Files to create:**
- `lib/fql/protocols/serializable.ex`
- `lib/fql/protocols/formattable.ex`
- `lib/fql/protocols/classifiable.ex`

**Estimated effort:** Medium-High

---

### 3. Use Elixir's Logger
**Current:** Custom logging with `IO.puts`
**Improvement:** Use Elixir's built-in Logger

```elixir
# Current
defp log(engine, message) do
  if engine.logger do
    IO.puts("[FQL Engine] #{message}")
  end
end

# Proposed
require Logger

defp log(engine, message, metadata \\ []) do
  if engine.logger do
    Logger.info(message, metadata)
  end
end

# Usage
log(engine, "Setting key-value", query: query, class: class)
```

**Benefits:**
- Structured logging
- Log levels (debug, info, warn, error)
- Metadata support
- Integration with logging backends

**Files to change:**
- `lib/fql/engine.ex`
- Add Logger configuration in `config/config.exs`

**Estimated effort:** Low

---

### 4. Add Telemetry Events
**Current:** No instrumentation
**Improvement:** Add telemetry for monitoring and metrics

```elixir
:telemetry.execute(
  [:fql, :engine, :set],
  %{duration: duration},
  %{class: class, directory: path}
)
```

**Events to add:**
- `[:fql, :engine, :set]` - Set operations
- `[:fql, :engine, :clear]` - Clear operations
- `[:fql, :engine, :read_single]` - Single reads
- `[:fql, :engine, :read_range]` - Range reads
- `[:fql, :parser, :parse]` - Parser operations

**Benefits:**
- Production monitoring
- Performance metrics
- Integration with observability tools

**Files to create:**
- `lib/fql/telemetry.ex` - Telemetry definitions
- Update `lib/fql/engine.ex` to emit events
- Update `lib/fql/parser.ex` to emit events

**Dependencies to add:**
- `{:telemetry, "~> 1.2"}`

**Estimated effort:** Medium

---

## Medium Priority - Production Readiness

### 5. OTP Application Structure
**Current:** Basic application with empty supervisor
**Improvement:** Proper OTP supervision tree

```elixir
defmodule Fql.Application do
  use Application

  def start(_type, _args) do
    children = [
      {Fql.ConnectionPool, []},
      {Fql.MetricsReporter, []},
      {Task.Supervisor, name: Fql.TaskSupervisor}
    ]

    opts = [strategy: :one_for_one, name: Fql.Supervisor]
    Supervisor.start_link(children, opts)
  end
end
```

**Components to add:**
- Connection pool for FDB connections
- Metrics reporter
- Task supervisor for async operations

**Files to create:**
- `lib/fql/connection_pool.ex`
- `lib/fql/metrics_reporter.ex`

**Estimated effort:** Medium-High

---

### 6. GenServer-based Engine
**Current:** Engine is a struct with functions
**Improvement:** Make Engine a GenServer for state management

```elixir
defmodule Fql.Engine do
  use GenServer

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  def set(query, opts \\ []) do
    GenServer.call(__MODULE__, {:set, query, opts})
  end

  # ... other operations
end
```

**Benefits:**
- State management
- Connection pooling
- Concurrent query handling
- Better error recovery

**Files to change:**
- `lib/fql/engine.ex` - Convert to GenServer
- Update tests to start GenServer

**Estimated effort:** High (breaking change)

---

### 7. Configuration Management
**Current:** Options passed at runtime
**Improvement:** Use Application configuration

```elixir
# config/config.exs
config :fql,
  byte_order: :big,
  cluster_file: "/etc/foundationdb/fdb.cluster",
  default_timeout: 5_000

# Usage
byte_order = Application.get_env(:fql, :byte_order, :big)
```

**Files to create:**
- `config/config.exs`
- `config/dev.exs`
- `config/test.exs`
- `config/prod.exs`

**Files to change:**
- `lib/fql/engine.ex` - Use config for defaults

**Estimated effort:** Low

---

### 8. Custom Exception Types
**Current:** String error messages in tuples
**Improvement:** Define custom exceptions

```elixir
defmodule Fql.ParseError do
  defexception [:message, :query, :position]

  def exception(opts) do
    query = Keyword.get(opts, :query)
    position = Keyword.get(opts, :position)
    message = "Parse error at position #{position}: #{query}"

    %__MODULE__{
      message: message,
      query: query,
      position: position
    }
  end
end

# Usage
raise Fql.ParseError, query: query, position: pos
```

**Exceptions to define:**
- `Fql.ParseError`
- `Fql.ClassificationError`
- `Fql.SerializationError`
- `Fql.DatabaseError`

**Files to create:**
- `lib/fql/exceptions.ex`

**Files to change:**
- All modules to use exceptions
- Update error handling in tests

**Estimated effort:** Medium

---

### 9. Macro-based Query DSL
**Current:** Manual query construction
**Improvement:** Provide a nice DSL using macros

```elixir
# Current
KeyVal.key_value(
  KeyVal.key(["test"], [KeyVal.int(1)]),
  "value"
)

# Proposed
import Fql.DSL

query do
  dir ["test"]
  tuple [int(1)]
  value "value"
end

# Or even nicer
fql "/test(1) = value"
```

**Files to create:**
- `lib/fql/dsl.ex`
- `test/fql/dsl_test.exs`

**Estimated effort:** Medium-High

---

## Medium Priority - Developer Experience

### 10. ExDoc Documentation
**Current:** Basic @moduledoc and @doc
**Improvement:** Comprehensive documentation with examples

```elixir
@moduledoc """
The FQL Engine executes queries against FoundationDB.

## Examples

    iex> engine = Engine.new(db)
    iex> query = KeyVal.key_value(...)
    iex> Engine.set(engine, query)
    :ok

## Options

- `:byte_order` - `:big` or `:little` (default: `:big`)
- `:logger` - Enable logging (default: `false`)
"""
```

**Add:**
- Module documentation with examples
- Function documentation with examples
- Guides in `guides/` directory
- API reference generation

**Files to create:**
- `guides/getting_started.md`
- `guides/query_language.md`
- `guides/advanced_usage.md`

**Dependencies to add:**
- `{:ex_doc, "~> 0.30", only: :dev, runtime: false}`

**Estimated effort:** Medium

---

### 11. Dialyzer Type Specifications
**Current:** Basic @spec annotations
**Improvement:** Complete type specifications with Dialyzer

```elixir
# Add to mix.exs
def project do
  [
    dialyzer: [
      plt_add_apps: [:ex_unit],
      flags: [:error_handling, :underspecs]
    ]
  ]
end
```

**Tasks:**
- Complete all @spec annotations
- Fix all Dialyzer warnings
- Add @typedoc for custom types
- Create PLT files for CI

**Dependencies to add:**
- `{:dialyxir, "~> 1.4", only: [:dev, :test], runtime: false}`

**Estimated effort:** Medium

---

### 12. Credo Code Quality
**Current:** No linting
**Improvement:** Add Credo for code quality

```elixir
# .credo.exs
%{
  configs: [
    %{
      name: "default",
      strict: true,
      checks: [
        {Credo.Check.Readability.ModuleDoc, []},
        {Credo.Check.Design.AliasUsage, false}
      ]
    }
  ]
}
```

**Dependencies to add:**
- `{:credo, "~> 1.7", only: [:dev, :test], runtime: false}`

**Estimated effort:** Low

---

## Low Priority - Performance & Testing

### 13. Property-based Testing
**Current:** Example-based tests
**Improvement:** Add property-based tests with StreamData

```elixir
property "parse and format are inverse operations" do
  check all query <- query_generator() do
    {:ok, parsed} = Parser.parse(query)
    formatted = Format.format(parsed)
    {:ok, reparsed} = Parser.parse(formatted)

    assert parsed == reparsed
  end
end
```

**Tests to add:**
- Parse/format round-trips
- Serialization round-trips
- Query classification properties
- Tuple ordering properties

**Dependencies to add:**
- `{:stream_data, "~> 0.6", only: :test}`

**Files to create:**
- `test/property_test.exs`
- `test/support/generators.ex`

**Estimated effort:** Medium

---

### 14. Benchmarking Suite
**Current:** No benchmarks
**Improvement:** Add Benchee benchmarks

```elixir
Benchee.run(%{
  "parse simple query" => fn ->
    Parser.parse("/test(1) = value")
  end,
  "parse complex query" => fn ->
    Parser.parse("/path/to/test(1, 2, true) = (nested, tuple)")
  end
})
```

**Benchmarks to add:**
- Parser performance
- Serialization performance
- Classification performance
- Format performance

**Dependencies to add:**
- `{:benchee, "~> 1.1", only: :dev}`

**Files to create:**
- `bench/parser_bench.exs`
- `bench/engine_bench.exs`
- `bench/serialization_bench.exs`

**Estimated effort:** Low-Medium

---

### 15. Stream Optimization
**Current:** Basic Stream.resource implementation
**Improvement:** Optimize for large datasets

```elixir
def stream_range(engine, query, opts \\ []) do
  Stream.unfold(initial_state, fn state ->
    case fetch_batch(state) do
      {:ok, [], _new_state} -> nil
      {:ok, results, new_state} -> {results, new_state}
      {:error, _reason} -> nil
    end
  end)
  |> Stream.flat_map(& &1)
end
```

**Improvements:**
- Configurable batch size
- Parallel fetching
- Backpressure handling
- Memory-efficient processing

**Files to change:**
- `lib/fql/engine.ex`

**Estimated effort:** Medium

---

### 16. NimbleParsec Integration
**Current:** Hand-written recursive descent parser
**Improvement:** Use NimbleParsec for better performance

```elixir
defmodule Fql.Parser.Combinators do
  import NimbleParsec

  directory =
    ignore(string("/"))
    |> utf8_string([not: ?/], min: 1)
    |> repeat()

  # ... more combinators
end
```

**Benefits:**
- Better performance
- More maintainable
- Better error messages
- Streaming parsing

**Dependencies to add:**
- `{:nimble_parsec, "~> 1.3"}`

**Files to change:**
- Complete rewrite of `lib/fql/parser.ex`
- Keep scanner or replace with NimbleParsec

**Estimated effort:** High (breaking change)

---

## Low Priority - Additional Features

### 17. Query Validation
**Current:** Runtime validation during execution
**Improvement:** Validate queries before execution

```elixir
defmodule Fql.Validator do
  def validate(query) do
    with :ok <- validate_structure(query),
         :ok <- validate_types(query),
         :ok <- validate_constraints(query) do
      :ok
    end
  end
end
```

**Validations:**
- Schema validation
- Type compatibility
- Constraint checking
- Security checks

**Files to create:**
- `lib/fql/validator.ex`
- `test/fql/validator_test.exs`

**Estimated effort:** Medium

---

### 18. Query Builder
**Current:** Manual construction
**Improvement:** Fluent query builder API

```elixir
Query.new()
|> Query.directory(["path", "to", "dir"])
|> Query.tuple([int(1), int(2)])
|> Query.value("test")
|> Query.build()
```

**Files to create:**
- `lib/fql/query_builder.ex`
- `test/fql/query_builder_test.exs`

**Estimated effort:** Medium

---

### 19. JSON Serialization
**Current:** No JSON support
**Improvement:** Serialize queries to/from JSON

```elixir
defimpl Jason.Encoder, for: Fql.KeyVal.KeyValue do
  def encode(kv, opts) do
    Jason.Encode.map(%{
      "key" => kv.key,
      "value" => kv.value
    }, opts)
  end
end
```

**Dependencies to add:**
- `{:jason, "~> 1.4"}`

**Use cases:**
- Debugging
- API responses
- Query storage
- Cross-language compatibility

**Files to create:**
- `lib/fql/json.ex`
- `test/fql/json_test.exs`

**Estimated effort:** Low-Medium

---

### 20. Query Caching
**Current:** No caching
**Improvement:** Cache parsed queries

```elixir
defmodule Fql.QueryCache do
  use GenServer

  def get_or_parse(query_string) do
    case :ets.lookup(:query_cache, query_string) do
      [{^query_string, parsed}] -> {:ok, parsed}
      [] ->
        {:ok, parsed} = Parser.parse(query_string)
        :ets.insert(:query_cache, {query_string, parsed})
        {:ok, parsed}
    end
  end
end
```

**Benefits:**
- Faster repeated queries
- Reduced CPU usage
- Better performance

**Files to create:**
- `lib/fql/query_cache.ex`

**Estimated effort:** Low-Medium

---

## Summary by Priority

### High Priority (Core Idioms)
1. ✅ Replace Maps with Structs - Better pattern matching
2. ✅ Implement Elixir Protocols - Extensibility
3. ✅ Use Elixir's Logger - Proper logging
4. ✅ Add Telemetry Events - Observability

### Medium Priority (Production)
5. ✅ OTP Application Structure - Fault tolerance
6. ✅ GenServer-based Engine - State management
7. ✅ Configuration Management - Environment-based config
8. ✅ Custom Exception Types - Better error handling
9. ✅ Macro-based Query DSL - Developer experience

### Medium Priority (Developer Experience)
10. ✅ ExDoc Documentation - Better docs
11. ✅ Dialyzer Type Specifications - Type safety
12. ✅ Credo Code Quality - Linting

### Low Priority (Performance & Testing)
13. ⭐ Property-based Testing - Better coverage
14. ⭐ Benchmarking Suite - Performance metrics
15. ⭐ Stream Optimization - Better performance
16. ⭐ NimbleParsec Integration - Parser performance

### Low Priority (Additional Features)
17. 💡 Query Validation - Safety
18. 💡 Query Builder - Fluent API
19. 💡 JSON Serialization - Interoperability
20. 💡 Query Caching - Performance

## Implementation Order

### Phase 1: Make it Idiomatic (Weeks 1-2)
- Structs (item 1)
- Logger (item 3)
- Configuration (item 7)
- ExDoc (item 10)

### Phase 2: Production Ready (Weeks 3-4)
- Protocols (item 2)
- Telemetry (item 4)
- Exceptions (item 8)
- OTP structure (item 5)

### Phase 3: Performance (Weeks 5-6)
- GenServer engine (item 6)
- Stream optimization (item 15)
- Benchmarks (item 14)
- Query caching (item 20)

### Phase 4: Polish (Weeks 7-8)
- Dialyzer (item 11)
- Credo (item 12)
- Property tests (item 13)
- DSL macros (item 9)

### Phase 5: Optional (Future)
- NimbleParsec (item 16)
- Query validation (item 17)
- Query builder (item 18)
- JSON support (item 19)

## Breaking Changes

These items require API changes:
- **Item 1** (Structs) - Changes from maps to structs
- **Item 6** (GenServer) - Changes synchronous to async API
- **Item 8** (Exceptions) - Changes from tuples to exceptions
- **Item 16** (NimbleParsec) - Parser rewrite

Recommend deprecation period for these changes.

## Non-Breaking Additions

These can be added without breaking existing code:
- Items 2, 3, 4, 7, 9, 10, 11, 12, 13, 14, 15, 17, 18, 19, 20

## Estimated Total Effort

- **Phase 1:** 2-3 weeks
- **Phase 2:** 3-4 weeks
- **Phase 3:** 2-3 weeks
- **Phase 4:** 2-3 weeks
- **Phase 5:** 4-6 weeks (optional)

**Total:** 13-19 weeks for all items (excluding Phase 5)
