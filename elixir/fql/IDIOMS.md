# Elixir Idioms & Best Practices for FQL

This document provides specific examples of how to make the FQL codebase more idiomatic Elixir.

## 1. Pattern Matching Instead of Case Statements

### Current Implementation
```elixir
def parse_tuple_element(state) do
  case current_token(state) do
    %{kind: :other, value: "nil"} ->
      {:ok, :nil, advance(state)}
    %{kind: :other, value: "true"} ->
      {:ok, true, advance(state)}
    %{kind: :other, value: "false"} ->
      {:ok, false, advance(state)}
    %{kind: :other, value: value} ->
      parse_number_or_string(value, advance(state))
    _ ->
      {:error, "unexpected token in tuple"}
  end
end
```

### Idiomatic Version
```elixir
# Use function head pattern matching
def parse_tuple_element(%{pos: pos, tokens: tokens} = state) when pos < length(tokens) do
  token = Enum.at(tokens, pos)
  parse_token(token, state)
end

defp parse_token(%{kind: :other, value: "nil"}, state),
  do: {:ok, :nil, advance(state)}

defp parse_token(%{kind: :other, value: "true"}, state),
  do: {:ok, true, advance(state)}

defp parse_token(%{kind: :other, value: "false"}, state),
  do: {:ok, false, advance(state)}

defp parse_token(%{kind: :other, value: value}, state),
  do: parse_number_or_string(value, advance(state))

defp parse_token(_token, _state),
  do: {:error, "unexpected token in tuple"}
```

**Benefits:**
- More declarative
- Easier to read
- Better pattern matching optimization
- Each clause is a single responsibility

---

## 2. With Statements for Error Handling Pipelines

### Current Implementation
```elixir
defp execute_set(engine, query, class) do
  with {:ok, path} <- Convert.directory_to_path(query.key.directory),
       {:ok, tuple_bytes} <- Tuple.pack(query.key.tuple),
       {:ok, value_bytes} <- Values.pack(query.value, engine.byte_order, class == :vstamp_val) do
    # ... implementation
  end
end
```

### Idiomatic Enhancement
```elixir
defp execute_set(engine, query, class) do
  with {:ok, path} <- Convert.directory_to_path(query.key.directory),
       {:ok, tuple_bytes} <- Tuple.pack(query.key.tuple),
       {:ok, value_bytes} <- Values.pack(query.value, engine.byte_order, class == :vstamp_val),
       :ok <- perform_fdb_set(engine.db, path, tuple_bytes, value_bytes, class) do
    log(engine, "Successfully set key-value", query: query)
    :ok
  else
    {:error, :directory_error} = err ->
      log(engine, "Failed to convert directory", error: err)
      err

    {:error, :pack_error} = err ->
      log(engine, "Failed to pack data", error: err)
      err

    {:error, reason} = err ->
      log(engine, "FDB operation failed", error: reason)
      err
  end
end
```

**Benefits:**
- Explicit error handling for each case
- Better logging per error type
- Clear error propagation

---

## 3. Pipe Operator for Transformations

### Current Implementation
```elixir
def format_directory(dir) when is_list(dir) do
  Enum.map_join(dir, fn elem ->
    "/" <> format_dir_element(elem)
  end)
end
```

### Idiomatic Version
```elixir
def format_directory(dir) when is_list(dir) do
  dir
  |> Enum.map(&format_dir_element/1)
  |> Enum.map_join("", &"/#{&1}")
end

# Or even more concise
def format_directory(dir) when is_list(dir) do
  dir
  |> Stream.map(&format_dir_element/1)
  |> Stream.map(&"/#{&1}")
  |> Enum.join()
end
```

**Benefits:**
- Data transformation flow is clear
- Left-to-right reading
- Easy to add/remove steps
- Can use Stream for lazy evaluation

---

## 4. Use Structs with Defaults

### Current Implementation
```elixir
defstruct [:db, :byte_order, :logger]

def new(db, opts \\ []) do
  %__MODULE__{
    db: db,
    byte_order: Keyword.get(opts, :byte_order, :big),
    logger: Keyword.get(opts, :logger, nil)
  }
end
```

### Idiomatic Version
```elixir
defstruct db: nil,
          byte_order: :big,
          logger: nil,
          timeout: 5_000

def new(db, opts \\ []) do
  struct!(__MODULE__, [db: db] ++ opts)
end

# Or with validation
def new(db, opts \\ []) do
  engine = struct!(__MODULE__, [db: db] ++ opts)
  validate_engine!(engine)
  engine
end

defp validate_engine!(%__MODULE__{byte_order: order} = engine)
  when order in [:big, :little] do
  engine
end

defp validate_engine!(%__MODULE__{byte_order: order}) do
  raise ArgumentError, "byte_order must be :big or :little, got: #{inspect(order)}"
end
```

**Benefits:**
- Default values in struct definition
- Validation at construction time
- Clear contract
- Better error messages

---

## 5. Protocols for Polymorphism

### Current Implementation
```elixir
def pack(val, order, vstamp) when is_boolean(val) do
  {:ok, if(val, do: <<1>>, else: <<0>>)}
end

def pack({:int, val}, order, _vstamp) when is_integer(val) do
  bytes = encode_int64(val, order)
  {:ok, bytes}
end

# ... many more clauses
```

### Idiomatic Version
```elixir
defprotocol Fql.Packable do
  @doc "Pack a value into bytes"
  @spec pack(t(), keyword()) :: {:ok, binary()} | {:error, String.t()}
  def pack(value, opts)
end

defimpl Fql.Packable, for: Integer do
  def pack(value, opts) do
    order = Keyword.get(opts, :byte_order, :big)
    bytes = encode_int64(value, order)
    {:ok, bytes}
  end

  defp encode_int64(val, :big), do: <<val::big-signed-64>>
  defp encode_int64(val, :little), do: <<val::little-signed-64>>
end

defimpl Fql.Packable, for: BitString do
  def pack(value, _opts), do: {:ok, value}
end

defimpl Fql.Packable, for: Atom do
  def pack(true, _opts), do: {:ok, <<1>>}
  def pack(false, _opts), do: {:ok, <<0>>}
  def pack(:nil, _opts), do: {:ok, <<>>}
  def pack(atom, _opts), do: {:error, "cannot pack atom: #{atom}"}
end

# Usage
Fql.Packable.pack(value, byte_order: :big)
```

**Benefits:**
- Extensible without modifying core
- Type-based dispatch
- Clear separation of concerns
- Can add new types externally

---

## 6. Guards for Type Checking

### Current Implementation
```elixir
def pack({:int, val}, order, _vstamp) when is_integer(val) do
  bytes = encode_int64(val, order)
  {:ok, bytes}
end
```

### More Idiomatic
```elixir
# Define custom guards
defguard is_fql_int(val) when is_tuple(val) and tuple_size(val) == 2 and
         elem(val, 0) == :int and is_integer(elem(val, 1))

defguard is_fql_uint(val) when is_tuple(val) and tuple_size(val) == 2 and
         elem(val, 0) == :uint and is_integer(elem(val, 1)) and elem(val, 1) >= 0

defguard is_valid_byte_order(order) when order in [:big, :little]

# Usage
def pack({:int, val} = int, order, _vstamp)
    when is_fql_int(int) and is_valid_byte_order(order) do
  bytes = encode_int64(val, order)
  {:ok, bytes}
end
```

**Benefits:**
- Reusable type checks
- Better compile-time optimization
- Self-documenting
- Pattern matching at function head

---

## 7. Comprehensions Over Enum.map

### Current Implementation
```elixir
defp convert_dir_elements([], acc) do
  {:ok, Enum.reverse(acc)}
end

defp convert_dir_elements([elem | rest], acc) do
  case elem do
    str when is_binary(str) ->
      convert_dir_elements(rest, [str | acc])
    %{types: _} ->
      {:error, "directory contains variable"}
    _ ->
      {:error, "invalid directory element"}
  end
end
```

### Idiomatic Version
```elixir
def directory_to_path(directory) do
  try do
    path = for elem <- directory do
      validate_dir_element!(elem)
    end
    {:ok, path}
  catch
    :error, reason -> {:error, reason}
  end
end

defp validate_dir_element!(str) when is_binary(str), do: str
defp validate_dir_element!(%{types: _}), do: throw({:error, "directory contains variable"})
defp validate_dir_element!(_), do: throw({:error, "invalid directory element"})

# Or without exceptions
def directory_to_path(directory) do
  directory
  |> Enum.reduce_while({:ok, []}, fn elem, {:ok, acc} ->
    case validate_dir_element(elem) do
      {:ok, val} -> {:cont, {:ok, [val | acc]}}
      {:error, _} = err -> {:halt, err}
    end
  end)
  |> case do
    {:ok, path} -> {:ok, Enum.reverse(path)}
    error -> error
  end
end
```

**Benefits:**
- More declarative
- Clearer intent
- Can use guards in comprehension

---

## 8. Access Behaviour for Nested Data

### Current Implementation
```elixir
def get_value_type(query) do
  query.value.types
end
```

### Idiomatic Version
```elixir
# If using structs with Access behaviour
def get_value_type(query) do
  get_in(query, [:value, :types])
end

# Or with fallback
def get_value_type(query) do
  get_in(query, [Access.key(:value), Access.key(:types, [])])
end

# Or pattern matching
def get_value_type(%{value: %{types: types}}), do: types
def get_value_type(%{value: _}), do: []
```

**Benefits:**
- Safe navigation
- Default values
- Clear data access patterns

---

## 9. Stream for Lazy Evaluation

### Current Implementation
```elixir
def stream_range(engine, query, opts) do
  Stream.resource(
    fn -> initialize(engine, query, opts) end,
    fn state -> fetch_next(state) end,
    fn state -> cleanup(state) end
  )
end
```

### Enhanced Idiomatic Version
```elixir
def stream_range(engine, query, opts \\ []) do
  Stream.unfold(initial_state(engine, query, opts), &fetch_batch/1)
  |> Stream.flat_map(& &1)
  |> Stream.take_while(&valid_result?/1)
end

defp fetch_batch(%{done: true} = _state), do: nil
defp fetch_batch(state) do
  case do_fetch(state) do
    {:ok, results, new_state} -> {results, new_state}
    {:error, _reason} -> nil
  end
end

# Can compose with other stream operations
def stream_range_filtered(engine, query, filter_fn, opts \\ []) do
  engine
  |> stream_range(query, opts)
  |> Stream.filter(filter_fn)
end

# Can add transformation
def stream_range_mapped(engine, query, map_fn, opts \\ []) do
  engine
  |> stream_range(query, opts)
  |> Stream.map(map_fn)
end
```

**Benefits:**
- Composable streams
- Lazy evaluation
- Memory efficient
- Functional composition

---

## 10. Use Kernel.SpecialForms for Better Code

### Current Implementation
```elixir
defp collect_string_contents(state, acc) do
  case current_token(state) do
    %{kind: :str_mark} -> {:ok, acc, advance(state)}
    %{kind: :escape, value: char} -> collect_string_contents(advance(state), acc <> char)
    %{kind: _, value: value} -> collect_string_contents(advance(state), acc <> value)
    nil -> {:error, "unterminated string"}
  end
end
```

### Idiomatic Version Using IO Lists
```elixir
defp collect_string_contents(state, acc \\ []) do
  case current_token(state) do
    %{kind: :str_mark} ->
      # Convert iolist to binary only at the end
      {:ok, IO.iodata_to_binary(Enum.reverse(acc)), advance(state)}

    %{kind: :escape, value: char} ->
      collect_string_contents(advance(state), [char | acc])

    %{kind: _, value: value} ->
      collect_string_contents(advance(state), [value | acc])

    nil ->
      {:error, "unterminated string"}
  end
end
```

**Benefits:**
- More efficient (no string concatenation in loop)
- Builds list, converts once
- Better memory usage

---

## 11. Behaviour for Extensibility

### Define a Behaviour
```elixir
defmodule Fql.Storage do
  @moduledoc """
  Behaviour for storage backends.
  """

  @callback set(key :: binary(), value :: binary(), opts :: keyword()) ::
    :ok | {:error, term()}

  @callback get(key :: binary(), opts :: keyword()) ::
    {:ok, binary()} | {:error, :not_found} | {:error, term()}

  @callback delete(key :: binary(), opts :: keyword()) ::
    :ok | {:error, term()}

  @callback get_range(start_key :: binary(), end_key :: binary(), opts :: keyword()) ::
    {:ok, [{binary(), binary()}]} | {:error, term()}
end

# Implement for FDB
defmodule Fql.Storage.FoundationDB do
  @behaviour Fql.Storage

  @impl true
  def set(key, value, opts) do
    # FDB implementation
  end

  @impl true
  def get(key, opts) do
    # FDB implementation
  end

  # ... other callbacks
end

# Implement for testing
defmodule Fql.Storage.Memory do
  @behaviour Fql.Storage

  # In-memory ETS-based implementation for testing
end
```

**Benefits:**
- Swappable backends
- Easy testing
- Clear contracts
- Compile-time callback checking

---

## 12. Use `use` and `__using__` for Code Injection

### Create a Module Template
```elixir
defmodule Fql.Query do
  @moduledoc """
  Common functionality for query modules.
  """

  defmacro __using__(opts) do
    quote do
      alias Fql.KeyVal
      alias Fql.KeyVal.{Class, Convert, Tuple, Values}

      require Logger

      @behaviour Fql.Queryable

      # Default implementations
      def validate(query) do
        with :ok <- validate_structure(query),
             :ok <- validate_types(query) do
          :ok
        end
      end

      defoverridable validate: 1

      # Imported from opts
      if Keyword.get(unquote(opts), :telemetry, false) do
        defp emit_telemetry(event, measurements, metadata) do
          :telemetry.execute([:fql | event], measurements, metadata)
        end
      end
    end
  end
end

# Usage
defmodule Fql.SetQuery do
  use Fql.Query, telemetry: true

  # Automatically gets all the imports and default functions
end
```

**Benefits:**
- Code reuse
- Consistent module structure
- Configurable imports
- Reduced boilerplate

---

## 13. Task for Async Operations

### Current Implementation
```elixir
def read_range(engine, query, opts) do
  # Synchronous read
end
```

### Async Version
```elixir
def read_range_async(engine, query, opts \\ []) do
  Task.async(fn ->
    read_range(engine, query, opts)
  end)
end

def read_range_await(task, timeout \\ 5000) do
  Task.await(task, timeout)
end

# Or use Task.Supervisor
def read_range_supervised(engine, query, opts \\ []) do
  Task.Supervisor.async(Fql.TaskSupervisor, fn ->
    read_range(engine, query, opts)
  end)
end

# Parallel reads
def read_ranges_parallel(engine, queries, opts \\ []) do
  queries
  |> Enum.map(&Task.async(fn -> read_range(engine, &1, opts) end))
  |> Enum.map(&Task.await(&1, opts[:timeout] || 5000))
end
```

**Benefits:**
- Concurrent operations
- Better resource utilization
- Timeout handling
- Supervised tasks for fault tolerance

---

## 14. Agent for State Management

### Simple Cache Implementation
```elixir
defmodule Fql.QueryCache do
  use Agent

  def start_link(_opts) do
    Agent.start_link(fn -> %{} end, name: __MODULE__)
  end

  def get_or_parse(query_string) do
    case Agent.get(__MODULE__, &Map.get(&1, query_string)) do
      nil ->
        {:ok, parsed} = Fql.Parser.parse(query_string)
        Agent.update(__MODULE__, &Map.put(&1, query_string, parsed))
        {:ok, parsed}

      cached ->
        {:ok, cached}
    end
  end

  def clear do
    Agent.update(__MODULE__, fn _ -> %{} end)
  end
end
```

**Benefits:**
- Simple state management
- Process isolation
- Easy testing
- No locking needed

---

## 15. Registry for Dynamic Processes

### Engine Pool with Registry
```elixir
defmodule Fql.EnginePool do
  def start_engine(name, db_config) do
    spec = {Fql.Engine, [db_config, name: via_tuple(name)]}

    DynamicSupervisor.start_child(Fql.EngineSupervisor, spec)
  end

  def get_engine(name) do
    case Registry.lookup(Fql.EngineRegistry, name) do
      [{pid, _}] -> {:ok, pid}
      [] -> {:error, :not_found}
    end
  end

  defp via_tuple(name) do
    {:via, Registry, {Fql.EngineRegistry, name}}
  end
end

# Usage
{:ok, _pid} = Fql.EnginePool.start_engine(:main_engine, db_config)
{:ok, engine} = Fql.EnginePool.get_engine(:main_engine)
```

**Benefits:**
- Named processes
- Dynamic process creation
- Process discovery
- Distributed ready

---

## Summary of Key Idioms

1. **Pattern matching in function heads** - Not case statements
2. **With statements** - For error handling pipelines
3. **Pipe operators** - For data transformation
4. **Structs with defaults** - Better than plain maps
5. **Protocols** - For polymorphic behavior
6. **Guards** - For type checking
7. **Comprehensions** - Instead of Enum.map when appropriate
8. **Access behaviour** - For safe nested access
9. **Streams** - For lazy evaluation
10. **IO lists** - For efficient string building
11. **Behaviours** - For extensibility contracts
12. **`use` and `__using__`** - For code injection
13. **Task** - For async operations
14. **Agent** - For simple state
15. **Registry** - For process discovery

## Migration Strategy

1. **Start small** - Pick one idiom to improve
2. **Test thoroughly** - Add tests before refactoring
3. **Incremental changes** - Don't refactor everything at once
4. **Benchmark** - Ensure performance doesn't degrade
5. **Document** - Update docs as you go

## Resources

- [Elixir Style Guide](https://github.com/christopheradams/elixir_style_guide)
- [Credo Rules](https://hexdocs.pm/credo/overview.html)
- [Elixir in Action](https://www.manning.com/books/elixir-in-action-second-edition)
- [The Little Elixir & OTP Guidebook](https://www.manning.com/books/the-little-elixir-and-otp-guidebook)
