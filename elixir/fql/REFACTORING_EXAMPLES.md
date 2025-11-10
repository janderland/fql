# Concrete Refactoring Examples

This document shows specific refactorings for the FQL codebase with before/after comparisons.

## Example 1: Engine Module - Structs and Protocols

### Before (Current)
```elixir
# lib/fql/engine.ex
defmodule Fql.Engine do
  alias Fql.KeyVal
  alias Fql.KeyVal.Class
  alias Fql.KeyVal.Values
  alias Fql.KeyVal.Convert
  alias Fql.KeyVal.Tuple

  defstruct [:db, :byte_order, :logger]

  @type t :: %__MODULE__{
    db: term(),
    byte_order: :big | :little,
    logger: term()
  }

  def new(db, opts \\ []) do
    %__MODULE__{
      db: db,
      byte_order: Keyword.get(opts, :byte_order, :big),
      logger: Keyword.get(opts, :logger, nil)
    }
  end

  def set(engine, query) do
    case Class.classify(query) do
      class when class in [:constant, :vstamp_key, :vstamp_val] ->
        execute_set(engine, query, class)
      {:invalid, reason} ->
        {:error, "invalid query: #{reason}"}
      class ->
        {:error, "invalid query class for set: #{class}"}
    end
  end
end
```

### After (Idiomatic)
```elixir
# lib/fql/engine.ex
defmodule Fql.Engine do
  @moduledoc """
  Executes FQL queries against FoundationDB.

  ## Examples

      iex> {:ok, engine} = Engine.start_link(db_config)
      iex> query = Query.set("/test(1)", "value")
      iex> Engine.execute(engine, query)
      :ok
  """

  use GenServer
  require Logger

  alias Fql.Query
  alias Fql.Telemetry

  defstruct [
    db: nil,
    byte_order: :big,
    timeout: 5_000,
    max_retries: 3
  ]

  @type t :: %__MODULE__{
    db: pid() | nil,
    byte_order: :big | :little,
    timeout: pos_integer(),
    max_retries: non_neg_integer()
  }

  # Client API

  def start_link(opts \\ []) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @spec execute(GenServer.server(), Query.t(), keyword()) ::
    :ok | {:ok, term()} | {:error, term()}
  def execute(server \\ __MODULE__, query, opts \\ []) do
    GenServer.call(server, {:execute, query, opts})
  end

  # Server Callbacks

  @impl true
  def init(opts) do
    state = struct!(__MODULE__, opts)
    {:ok, state, {:continue, :connect}}
  end

  @impl true
  def handle_continue(:connect, state) do
    case connect_to_db(state) do
      {:ok, db} ->
        {:noreply, %{state | db: db}}
      {:error, reason} ->
        Logger.error("Failed to connect to FDB: #{inspect(reason)}")
        {:stop, reason, state}
    end
  end

  @impl true
  def handle_call({:execute, query, opts}, _from, state) do
    start_time = System.monotonic_time()

    result = with {:ok, classified} <- Query.classify(query),
                  {:ok, result} <- do_execute(classified, query, state, opts) do
      emit_telemetry(:success, query, start_time)
      {:ok, result}
    else
      {:error, reason} = error ->
        emit_telemetry(:error, query, start_time, reason)
        error
    end

    {:reply, result, state}
  end

  # Private Functions

  defp do_execute(:set, query, state, opts) do
    Query.Set.execute(query, state.db, opts)
  end

  defp do_execute(:clear, query, state, opts) do
    Query.Clear.execute(query, state.db, opts)
  end

  defp do_execute(:read_single, query, state, opts) do
    Query.ReadSingle.execute(query, state.db, opts)
  end

  defp do_execute(:read_range, query, state, opts) do
    Query.ReadRange.execute(query, state.db, opts)
  end

  defp emit_telemetry(status, query, start_time, metadata \\ %{}) do
    duration = System.monotonic_time() - start_time

    Telemetry.emit(
      [:fql, :engine, :execute, status],
      %{duration: duration},
      Map.merge(%{query: query}, metadata)
    )
  end
end
```

**Benefits:**
- GenServer for state management
- Better separation of concerns
- Telemetry integration
- Proper OTP structure
- Type specs throughout
- Better error handling

---

## Example 2: Parser Module - NimbleParsec

### Before (Current)
```elixir
# lib/fql/parser.ex - Recursive descent parser
defmodule Fql.Parser do
  alias Fql.KeyVal
  alias Fql.Parser.Scanner

  def parse(query) do
    with {:ok, tokens} <- Scanner.scan(query),
         {:ok, result} <- parse_tokens(tokens) do
      {:ok, result}
    end
  end

  defp parse_tokens(tokens) do
    tokens = Enum.reject(tokens, fn t -> t.kind in [:whitespace, :newline] end)
    state = %{tokens: tokens, pos: 0, directory: [], tuple: [], value: nil}
    parse_query(state)
  end

  # ... 200+ lines of recursive descent parsing
end
```

### After (NimbleParsec)
```elixir
# lib/fql/parser.ex
defmodule Fql.Parser do
  @moduledoc """
  FQL query parser using NimbleParsec.
  """

  import NimbleParsec

  # Primitives
  whitespace = ascii_string([?\s, ?\t, ?\n, ?\r], min: 1)
  digit = ascii_char([?0..?9])
  letter = ascii_char([?a..?z, ?A..?Z])

  # Directory
  dir_sep = string("/")
  dir_name = utf8_string([not: ?/, not: ?<, not: ?(], min: 1)

  directory_element =
    choice([
      dir_name |> unwrap_and_tag(:dir_name),
      ignore(string("<"))
      |> concat(variable_types())
      |> ignore(string(">"))
      |> tag(:variable)
    ])

  directory =
    ignore(dir_sep)
    |> repeat(
      directory_element
      |> optional(ignore(dir_sep))
    )
    |> tag(:directory)

  # Tuple
  tuple_element =
    choice([
      string("nil") |> replace(:nil),
      string("true") |> replace(true),
      string("false") |> replace(false),
      integer_parser(),
      float_parser(),
      string_parser(),
      hex_parser(),
      variable_parser(),
      nested_tuple()
    ])

  tuple =
    ignore(string("("))
    |> optional(ignore(whitespace))
    |> optional(
      tuple_element
      |> repeat(
        ignore(string(","))
        |> optional(ignore(whitespace))
        |> concat(tuple_element)
      )
    )
    |> optional(ignore(whitespace))
    |> ignore(string(")"))
    |> tag(:tuple)

  # Value
  value =
    choice([
      string("clear") |> replace(:clear),
      variable_parser(),
      tuple,
      tuple_element
    ])

  # Complete query
  query =
    directory
    |> optional(tuple)
    |> optional(
      ignore(whitespace)
      |> ignore(string("="))
      |> ignore(whitespace)
      |> concat(value)
    )
    |> eos()

  defparsec :parse_query, query

  # Public API
  def parse(input) when is_binary(input) do
    case parse_query(input) do
      {:ok, ast, "", _, _, _} ->
        {:ok, build_query(ast)}

      {:ok, _, rest, _, _, _} ->
        {:error, "unexpected input: #{rest}"}

      {:error, reason, _rest, _, _, _} ->
        {:error, reason}
    end
  end

  # Helper parsers
  defp integer_parser do
    optional(string("-"))
    |> concat(integer(min: 1))
    |> reduce({List, :to_integer, []})
    |> unwrap_and_tag(:int)
  end

  defp float_parser do
    optional(string("-"))
    |> concat(integer(min: 1))
    |> concat(string("."))
    |> concat(integer(min: 1))
    |> reduce({List, :to_string, []})
    |> map({Float, :parse, []})
    |> unwrap_and_tag(:float)
  end

  # AST to query conversion
  defp build_query(ast) do
    # Convert parsed AST to Fql.Query structures
  end
end
```

**Benefits:**
- Better performance
- Declarative grammar
- Better error messages
- Less code
- Streaming support
- Parser combinators

---

## Example 3: KeyVal Module - Proper Structs

### Before (Current)
```elixir
# lib/fql/keyval.ex
defmodule Fql.KeyVal do
  @type key_value :: %{
    key: key(),
    value: value()
  }

  @type key :: %{
    directory: directory(),
    tuple: tuple()
  }

  def key_value(key, value) do
    %{key: key, value: value}
  end

  def key(directory, tuple) do
    %{directory: directory, tuple: tuple}
  end
end
```

### After (Idiomatic)
```elixir
# lib/fql/keyval.ex
defmodule Fql.KeyVal do
  @moduledoc """
  Core data structures for FQL queries.
  """

  # KeyValue struct
  defmodule KeyValue do
    @moduledoc """
    Represents a key-value pair in FQL.
    """

    @type t :: %__MODULE__{
      key: Fql.KeyVal.Key.t(),
      value: Fql.KeyVal.value()
    }

    @enforce_keys [:key, :value]
    defstruct [:key, :value]

    @doc """
    Creates a new KeyValue.

    ## Examples

        iex> key = Key.new(["test"], [int(1)])
        iex> KeyValue.new(key, "value")
        %KeyValue{key: key, value: "value"}
    """
    def new(key, value) do
      %__MODULE__{key: key, value: value}
    end

    defimpl String.Chars do
      def to_string(kv) do
        Fql.Parser.Format.format(kv)
      end
    end

    defimpl Inspect do
      def inspect(kv, opts) do
        Inspect.Algebra.concat([
          "#KeyValue<",
          Fql.Parser.Format.format(kv),
          ">"
        ])
      end
    end
  end

  # Key struct
  defmodule Key do
    @type t :: %__MODULE__{
      directory: Fql.KeyVal.directory(),
      tuple: Fql.KeyVal.tuple()
    }

    @enforce_keys [:directory, :tuple]
    defstruct [:directory, :tuple]

    def new(directory, tuple \\ []) do
      %__MODULE__{directory: directory, tuple: tuple}
    end

    defimpl String.Chars do
      def to_string(key) do
        Fql.Parser.Format.format_key(key)
      end
    end
  end

  # Variable struct
  defmodule Variable do
    @type t :: %__MODULE__{
      types: [Fql.KeyVal.value_type()]
    }

    defstruct types: []

    def new(types \\ []) do
      %__MODULE__{types: types}
    end

    def any, do: new([])
    def int, do: new([:int])
    def string, do: new([:string])
  end

  # Convenience functions
  def key_value(key, value), do: KeyValue.new(key, value)
  def key(directory, tuple \\ []), do: Key.new(directory, tuple)
  def variable(types \\ []), do: Variable.new(types)

  # Type helpers
  def int(value), do: {:int, value}
  def uint(value), do: {:uint, value}
  def bytes(value), do: {:bytes, value}
  def uuid(value), do: {:uuid, value}
end
```

**Benefits:**
- Enforced keys
- Better introspection
- Custom inspect/string protocols
- Namespaced types
- Self-documenting

---

## Example 4: Error Handling - Custom Exceptions

### Before (Current)
```elixir
# lib/fql/parser.ex
def parse_string(state) do
  collect_string_contents(state, "")
end

defp collect_string_contents(state, acc) do
  case current_token(state) do
    %{kind: :str_mark} -> {:ok, acc, advance(state)}
    nil -> {:error, "unterminated string"}
    # ...
  end
end
```

### After (With Exceptions)
```elixir
# lib/fql/exceptions.ex
defmodule Fql.ParseError do
  @moduledoc """
  Raised when a query string cannot be parsed.
  """

  defexception [:message, :query, :line, :column, :context]

  @type t :: %__MODULE__{
    message: String.t(),
    query: String.t(),
    line: non_neg_integer(),
    column: non_neg_integer(),
    context: map()
  }

  def exception(opts) do
    query = Keyword.fetch!(opts, :query)
    line = Keyword.get(opts, :line, 1)
    column = Keyword.get(opts, :column, 1)
    reason = Keyword.get(opts, :reason, "parse error")

    message = """
    Parse error on line #{line}, column #{column}: #{reason}

    #{highlight_error(query, line, column)}
    """

    %__MODULE__{
      message: message,
      query: query,
      line: line,
      column: column,
      context: Keyword.get(opts, :context, %{})
    }
  end

  defp highlight_error(query, line, column) do
    lines = String.split(query, "\n")
    error_line = Enum.at(lines, line - 1, "")

    """
    #{error_line}
    #{String.duplicate(" ", column - 1)}^
    """
  end
end

# lib/fql/parser.ex
def parse!(query) when is_binary(query) do
  case parse(query) do
    {:ok, result} -> result
    {:error, reason} -> raise Fql.ParseError, query: query, reason: reason
  end
end

def parse(query) when is_binary(query) do
  try do
    do_parse(query)
  rescue
    e in Fql.ParseError -> {:error, Exception.message(e)}
  end
end
```

**Benefits:**
- Better error messages
- Stack traces
- Error context
- Bang (!) variants
- Structured error data

---

## Example 5: Configuration - Application Environment

### Before (Current)
```elixir
# Hardcoded defaults everywhere
def new(db, opts \\ []) do
  %__MODULE__{
    db: db,
    byte_order: Keyword.get(opts, :byte_order, :big),
    logger: Keyword.get(opts, :logger, nil)
  }
end
```

### After (With Config)
```elixir
# config/config.exs
import Config

config :fql,
  byte_order: :big,
  default_timeout: 5_000,
  max_retries: 3,
  log_level: :info

import_config "#{config_env()}.exs"

# config/dev.exs
import Config

config :fql,
  log_level: :debug

# config/test.exs
import Config

config :fql,
  log_level: :warn,
  default_timeout: 1_000

# config/prod.exs
import Config

config :fql,
  log_level: :info,
  default_timeout: 10_000,
  max_retries: 5

# lib/fql/config.ex
defmodule Fql.Config do
  @moduledoc """
  Configuration helpers for FQL.
  """

  @doc """
  Gets a configuration value.
  """
  def get(key, default \\ nil) do
    Application.get_env(:fql, key, default)
  end

  @doc """
  Gets the byte order setting.
  """
  def byte_order do
    get(:byte_order, :big)
  end

  @doc """
  Gets the default timeout.
  """
  def default_timeout do
    get(:default_timeout, 5_000)
  end

  @doc """
  Validates configuration on application start.
  """
  def validate! do
    validate_byte_order!()
    validate_timeout!()
    :ok
  end

  defp validate_byte_order! do
    case byte_order() do
      order when order in [:big, :little] -> :ok
      other -> raise ArgumentError, "Invalid byte_order: #{inspect(other)}"
    end
  end

  defp validate_timeout! do
    case default_timeout() do
      timeout when is_integer(timeout) and timeout > 0 -> :ok
      other -> raise ArgumentError, "Invalid timeout: #{inspect(other)}"
    end
  end
end

# lib/fql/application.ex
defmodule Fql.Application do
  use Application

  def start(_type, _args) do
    # Validate config on startup
    Fql.Config.validate!()

    children = [
      # ... supervisors
    ]

    Supervisor.start_link(children, strategy: :one_for_one)
  end
end

# lib/fql/engine.ex
def init(opts) do
  state = %__MODULE__{
    db: Keyword.get(opts, :db),
    byte_order: Keyword.get(opts, :byte_order, Fql.Config.byte_order()),
    timeout: Keyword.get(opts, :timeout, Fql.Config.default_timeout())
  }

  {:ok, state}
end
```

**Benefits:**
- Environment-specific config
- Central configuration
- Validation on startup
- Runtime config possible
- Twelve-factor app compliant

---

## Example 6: Testing - Mocks and Stubs

### Before (Current)
```elixir
# Tests with nil db
setup do
  engine = Engine.new(nil, byte_order: :big)
  {:ok, engine: engine}
end
```

### After (With Mox)
```elixir
# test/support/mocks.ex
Mox.defmock(Fql.Storage.Mock, for: Fql.Storage)

# test/test_helper.exs
ExUnit.start()
Application.put_env(:fql, :storage_adapter, Fql.Storage.Mock)

# test/fql/engine_test.exs
defmodule Fql.EngineTest do
  use ExUnit.Case, async: true

  import Mox

  alias Fql.Engine
  alias Fql.Storage.Mock

  setup :verify_on_exit!

  describe "set/2" do
    test "calls storage adapter with correct parameters" do
      query = build_set_query()

      expect(Mock, :set, fn key, value, _opts ->
        assert key == expected_key()
        assert value == expected_value()
        :ok
      end)

      assert :ok = Engine.set(query)
    end

    test "handles storage errors" do
      query = build_set_query()

      expect(Mock, :set, fn _key, _value, _opts ->
        {:error, :connection_failed}
      end)

      assert {:error, :connection_failed} = Engine.set(query)
    end
  end

  # Test helpers
  defp build_set_query do
    # ...
  end
end

# lib/fql/engine.ex
defmodule Fql.Engine do
  @storage_adapter Application.compile_env(:fql, :storage_adapter, Fql.Storage.FoundationDB)

  defp perform_set(key, value, opts) do
    @storage_adapter.set(key, value, opts)
  end
end
```

**Benefits:**
- Proper mocking
- Async tests
- No real DB needed
- Behavior verification
- Test isolation

---

## Migration Path

### Phase 1: Non-Breaking Changes (Week 1-2)
1. Add configuration system
2. Add custom exceptions (keep tuple errors)
3. Add logging with Logger
4. Add telemetry events
5. Write comprehensive docs

### Phase 2: Additive Changes (Week 3-4)
6. Add GenServer version of Engine (keep old API)
7. Add protocols alongside current implementation
8. Add better structs (keep convenience functions)
9. Add behaviours for extensibility

### Phase 3: Breaking Changes (Week 5-6)
10. Switch to structs by default
11. Use GenServer as primary API
12. Deprecate old functions
13. Update all tests

### Phase 4: Polish (Week 7-8)
14. NimbleParsec migration
15. Complete protocol migration
16. Performance optimization
17. Production hardening

## Summary

Each refactoring:
- ✅ Improves code quality
- ✅ Makes code more idiomatic
- ✅ Adds better error handling
- ✅ Improves testability
- ✅ Maintains or improves performance
- ✅ Follows Elixir conventions

Start with the non-breaking changes to get immediate benefits without risk!
