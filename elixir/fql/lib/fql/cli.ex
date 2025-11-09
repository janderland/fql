defmodule Fql.CLI do
  @moduledoc """
  Command-line interface for FQL.

  This module provides the main entry point for the FQL CLI application.
  """

  alias Fql.Engine
  alias Fql.Parser

  @version "0.1.0"

  @doc """
  Main entry point for the escript.
  """
  def main(args) do
    case parse_args(args) do
      {:ok, opts} ->
        run(opts)

      {:error, message} ->
        IO.puts(:stderr, "Error: #{message}")
        print_usage()
        System.halt(1)
    end
  end

  @doc """
  Runs the FQL CLI with the given options.
  """
  def run(opts) do
    case opts.mode do
      :interactive ->
        run_interactive(opts)

      :query ->
        run_query(opts)

      :version ->
        IO.puts("FQL v#{@version}")
        :ok

      :help ->
        print_usage()
        :ok
    end
  end

  defp run_interactive(opts) do
    IO.puts("FQL Interactive Mode v#{@version}")
    IO.puts("Type 'exit' to quit, 'help' for help")
    IO.puts("")

    engine = initialize_engine(opts)
    interactive_loop(engine, opts)
  end

  defp interactive_loop(engine, opts) do
    case IO.gets("fql> ") do
      :eof ->
        IO.puts("\nGoodbye!")
        :ok

      {:error, reason} ->
        IO.puts(:stderr, "Input error: #{inspect(reason)}")
        System.halt(1)

      line ->
        line = String.trim(line)

        case line do
          "" ->
            interactive_loop(engine, opts)

          "exit" ->
            IO.puts("Goodbye!")
            :ok

          "help" ->
            print_help()
            interactive_loop(engine, opts)

          query_string ->
            execute_query_string(engine, query_string, opts)
            interactive_loop(engine, opts)
        end
    end
  end

  defp run_query(opts) do
    engine = initialize_engine(opts)

    Enum.each(opts.queries, fn query_string ->
      execute_query_string(engine, query_string, opts)
    end)
  end

  defp execute_query_string(engine, query_string, opts) do
    case Parser.parse(query_string) do
      {:ok, query} ->
        execute_query(engine, query, opts)

      {:error, reason} ->
        IO.puts(:stderr, "Parse error: #{reason}")
    end
  end

  defp execute_query(engine, query, opts) do
    # Classify and execute based on query type
    result = case query do
      %{key: _, value: _} ->
        # KeyValue query
        case Fql.KeyVal.Class.classify(query) do
          class when class in [:constant, :vstamp_key, :vstamp_val] ->
            if opts.write do
              Engine.set(engine, query)
            else
              {:error, "write queries not allowed without --write flag"}
            end

          :clear ->
            if opts.write do
              Engine.clear(engine, query)
            else
              {:error, "write queries not allowed without --write flag"}
            end

          :read_single ->
            Engine.read_single(engine, query, %{filter: opts.strict})

          :read_range ->
            Engine.read_range(engine, query, %{
              reverse: opts.reverse,
              filter: opts.strict,
              limit: opts.limit
            })

          {:invalid, reason} ->
            {:error, "invalid query: #{reason}"}

          other ->
            {:error, "unsupported query class: #{inspect(other)}"}
        end

      directory when is_list(directory) ->
        # Directory query
        Engine.list_directory(engine, directory)

      _ ->
        {:error, "unknown query type"}
    end

    case result do
      :ok ->
        IO.puts("OK")

      {:ok, data} ->
        print_result(data, opts)

      {:error, reason} ->
        IO.puts(:stderr, "Error: #{reason}")
    end
  end

  defp print_result(data, opts) when is_list(data) do
    Enum.each(data, fn item ->
      print_result(item, opts)
    end)
  end

  defp print_result(data, _opts) do
    IO.inspect(data, pretty: true, limit: :infinity)
  end

  defp initialize_engine(opts) do
    # In a real implementation, this would connect to FDB
    # db = :erlfdb.open(opts.cluster)
    db = nil

    Engine.new(db,
      byte_order: if(opts.little_endian, do: :little, else: :big),
      logger: if(opts.log, do: true, else: nil)
    )
  end

  defp parse_args(args) do
    {parsed, remaining, errors} = OptionParser.parse(args,
      strict: [
        help: :boolean,
        version: :boolean,
        query: :keep,
        cluster: :string,
        write: :boolean,
        little: :boolean,
        reverse: :boolean,
        strict: :boolean,
        bytes: :boolean,
        log: :boolean,
        log_file: :string,
        limit: :integer
      ],
      aliases: [
        h: :help,
        v: :version,
        q: :query,
        c: :cluster,
        w: :write,
        l: :little,
        r: :reverse,
        s: :strict,
        b: :bytes
      ]
    )

    cond do
      errors != [] ->
        {:error, "Invalid arguments: #{inspect(errors)}"}

      Keyword.get(parsed, :help, false) ->
        {:ok, %{mode: :help}}

      Keyword.get(parsed, :version, false) ->
        {:ok, %{mode: :version}}

      remaining != [] ->
        {:error, "Unexpected arguments: #{inspect(remaining)}"}

      true ->
        queries = Keyword.get_values(parsed, :query)

        mode = if queries == [], do: :interactive, else: :query

        {:ok, %{
          mode: mode,
          queries: queries,
          cluster: Keyword.get(parsed, :cluster, ""),
          write: Keyword.get(parsed, :write, false),
          little_endian: Keyword.get(parsed, :little, false),
          reverse: Keyword.get(parsed, :reverse, false),
          strict: Keyword.get(parsed, :strict, false),
          bytes: Keyword.get(parsed, :bytes, false),
          log: Keyword.get(parsed, :log, false),
          log_file: Keyword.get(parsed, :log_file, "log.txt"),
          limit: Keyword.get(parsed, :limit, 0)
        }}
    end
  end

  defp print_usage do
    IO.puts("""
    FQL - A query language for FoundationDB

    Usage:
      fql [flags] [query ...]

    Flags:
      -h, --help                Show this help message
      -v, --version             Show version
      -q, --query QUERY         Execute query non-interactively
      -c, --cluster FILE        Path to cluster file
      -w, --write               Allow write queries
      -l, --little              Use little endian encoding
      -r, --reverse             Query range-reads in reverse order
      -s, --strict              Throw error if KV doesn't match schema
      -b, --bytes               Print full byte strings
          --log                 Enable debug logging
          --log-file FILE       Logging file (default: log.txt)
          --limit N             Limit number of KVs in range-reads

    Examples:
      # Interactive mode
      fql

      # Execute a query
      fql -q "/my/dir(1, 2, 3) = value"

      # Read a range
      fql -q "/my/dir(<int>) = <>"

      # List directories
      fql -q "/my/dir"
    """)
  end

  defp print_help do
    IO.puts("""
    FQL Query Language:

    Directory queries:
      /path/to/dir              List all subdirectories

    Key-value queries:
      /dir(key) = value         Set a key-value
      /dir(key) = clear         Clear a key
      /dir(key)                 Read a single key
      /dir(<type>)              Read range of keys

    Variables:
      <>                        Match any value
      <int>                     Match integer values
      <string>                  Match string values
      <int|uint>                Match int or uint values

    Data types:
      123                       Integer
      "hello"                   String
      true, false               Boolean
      0xFF                      Bytes (hex)
      nil                       Null value

    Commands:
      help                      Show this help
      exit                      Exit interactive mode
    """)
  end
end
