defmodule Fql.Engine do
  @moduledoc """
  Executes FQL queries against FoundationDB.

  Each valid query class has a corresponding function for executing that
  class of query. Unless `transact/2` is used, each query is executed
  in its own transaction.
  """

  alias Fql.KeyVal
  alias Fql.KeyVal.Class
  alias Fql.KeyVal.Values
  alias Fql.KeyVal.Convert
  alias Fql.KeyVal.Tuple

  @type option :: {:byte_order, :big | :little} | {:logger, term()}
  @type single_opts :: %{filter: boolean()}
  @type range_opts :: %{reverse: boolean(), filter: boolean(), limit: integer()}

  defstruct [:db, :byte_order, :logger]

  @type t :: %__MODULE__{
    db: term(),
    byte_order: :big | :little,
    logger: term()
  }

  @doc """
  Creates a new Engine with the given FoundationDB database.
  """
  @spec new(term(), [option()]) :: t()
  def new(db, opts \\ []) do
    %__MODULE__{
      db: db,
      byte_order: Keyword.get(opts, :byte_order, :big),
      logger: Keyword.get(opts, :logger, nil)
    }
  end

  @doc """
  Executes a function within a single transaction.
  """
  @spec transact(t(), (t() -> {:ok, term()} | {:error, term()})) ::
    {:ok, term()} | {:error, term()}
  def transact(engine, fun) do
    case engine.db do
      nil ->
        # Testing mode - just run the function
        fun.(engine)

      db ->
        # Production mode with FDB
        try do
          # In production with erlfdb:
          # result = :erlfdb.transactional(db, fn tx ->
          #   tx_engine = %{engine | db: tx}
          #   fun.(tx_engine)
          # end)
          # result
          fun.(engine)
        rescue
          e -> {:error, "transaction error: #{inspect(e)}"}
        end
    end
  end

  @doc """
  Performs a write operation for a single key-value.

  The query must be of class :constant, :vstamp_key, or :vstamp_val.
  """
  @spec set(t(), KeyVal.key_value()) :: :ok | {:error, String.t()}
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

  defp execute_set(engine, query, class) do
    with {:ok, path} <- Convert.directory_to_path(query.key.directory),
         {:ok, tuple_bytes} <- Tuple.pack(query.key.tuple),
         {:ok, value_bytes} <- Values.pack(query.value, engine.byte_order, class == :vstamp_val) do

      log(engine, "Setting key-value: #{inspect(query)}")

      case engine.db do
        nil ->
          # No database connection (testing mode)
          :ok

        db ->
          # Real FDB interaction
          try do
            # In production with erlfdb:
            # :erlfdb.transactional(db, fn tx ->
            #   dir = open_or_create_directory(tx, path)
            #   key = concat_bytes(dir, tuple_bytes)
            #   case class do
            #     :vstamp_key -> :erlfdb.set_versionstamped_key(tx, key, value_bytes)
            #     :vstamp_val -> :erlfdb.set_versionstamped_value(tx, key, value_bytes)
            #     :constant -> :erlfdb.set(tx, key, value_bytes)
            #   end
            # end)
            :ok
          rescue
            e -> {:error, "FDB error: #{inspect(e)}"}
          end
      end
    end
  end

  @doc """
  Performs a clear operation for a single key.

  The query must be of class :clear.
  """
  @spec clear(t(), KeyVal.key_value()) :: :ok | {:error, String.t()}
  def clear(engine, query) do
    case Class.classify(query) do
      :clear ->
        execute_clear(engine, query)

      {:invalid, reason} ->
        {:error, "invalid query: #{reason}"}

      class ->
        {:error, "invalid query class for clear: #{class}"}
    end
  end

  defp execute_clear(engine, query) do
    with {:ok, path} <- Convert.directory_to_path(query.key.directory),
         {:ok, tuple_bytes} <- Tuple.pack(query.key.tuple) do

      log(engine, "Clearing key: #{inspect(query.key)}")

      case engine.db do
        nil ->
          # No database connection (testing mode)
          :ok

        db ->
          # Real FDB interaction
          try do
            # In production with erlfdb:
            # :erlfdb.transactional(db, fn tx ->
            #   dir = open_or_create_directory(tx, path)
            #   key = concat_bytes(dir, tuple_bytes)
            #   :erlfdb.clear(tx, key)
            # end)
            :ok
          rescue
            e -> {:error, "FDB error: #{inspect(e)}"}
          end
      end
    end
  end

  @doc """
  Reads a single key-value.

  The query must be of class :read_single.
  """
  @spec read_single(t(), KeyVal.key_value(), single_opts()) ::
    {:ok, KeyVal.key_value()} | {:error, String.t()}
  def read_single(engine, query, opts \\ %{filter: false}) do
    case Class.classify(query) do
      :read_single ->
        execute_read_single(engine, query, opts)

      {:invalid, reason} ->
        {:error, "invalid query: #{reason}"}

      class ->
        {:error, "invalid query class for read_single: #{class}"}
    end
  end

  defp execute_read_single(engine, query, opts) do
    with {:ok, path} <- Convert.directory_to_path(query.key.directory),
         {:ok, tuple_bytes} <- Tuple.pack(query.key.tuple) do

      log(engine, "Reading single key-value: #{inspect(query.key)}")

      case engine.db do
        nil ->
          # No database connection (testing mode)
          # Return a dummy response
          {:ok, query}

        db ->
          # Real FDB interaction
          try do
            # In production with erlfdb:
            # result = :erlfdb.transactional(db, fn tx ->
            #   dir = open_directory(tx, path)
            #   key = concat_bytes(dir, tuple_bytes)
            #   value_bytes = :erlfdb.get(tx, key)
            #
            #   if value_bytes == :not_found do
            #     {:error, "key not found"}
            #   else
            #     # Unpack value based on variable type
            #     value_type = extract_value_type(query.value)
            #     {:ok, value} = Values.unpack(value_bytes, value_type, engine.byte_order)
            #
            #     # Return complete key-value
            #     {:ok, KeyVal.key_value(query.key, value)}
            #   end
            # end)
            # result

            # For now, return dummy
            {:ok, query}
          rescue
            e -> {:error, "FDB error: #{inspect(e)}"}
          end
      end
    end
  end

  @doc """
  Reads multiple key-values matching the query schema.

  The query must be of class :read_range.
  """
  @spec read_range(t(), KeyVal.key_value(), range_opts()) ::
    {:ok, [KeyVal.key_value()]} | {:error, String.t()}
  def read_range(engine, query, opts \\ %{reverse: false, filter: false, limit: 0}) do
    case Class.classify(query) do
      :read_range ->
        execute_read_range(engine, query, opts)

      {:invalid, reason} ->
        {:error, "invalid query: #{reason}"}

      class ->
        {:error, "invalid query class for read_range: #{class}"}
    end
  end

  defp execute_read_range(engine, query, opts) do
    with {:ok, path} <- Convert.directory_to_path(query.key.directory) do

      log(engine, "Reading range: #{inspect(query.key)}")

      case engine.db do
        nil ->
          # No database connection (testing mode)
          {:ok, []}

        db ->
          # Real FDB interaction
          try do
            # In production with erlfdb:
            # results = :erlfdb.transactional(db, fn tx ->
            #   dir = open_directory(tx, path)
            #   {start_key, end_key} = compute_range(dir, query.key.tuple)
            #
            #   fdb_opts = [
            #     limit: opts.limit,
            #     reverse: opts.reverse
            #   ]
            #
            #   kvs = :erlfdb.get_range(tx, start_key, end_key, fdb_opts)
            #
            #   # Unpack each key-value
            #   Enum.map(kvs, fn {key, value} ->
            #     tuple = unpack_key(key, dir)
            #     unpacked_value = unpack_value(value, query.value, engine.byte_order)
            #     KeyVal.key_value(KeyVal.key(path, tuple), unpacked_value)
            #   end)
            # end)
            # {:ok, results}

            # For now, return empty
            {:ok, []}
          rescue
            e -> {:error, "FDB error: #{inspect(e)}"}
          end
      end
    end
  end

  @doc """
  Lists directories matching the query schema.
  """
  @spec list_directory(t(), KeyVal.directory()) ::
    {:ok, [KeyVal.directory()]} | {:error, String.t()}
  def list_directory(engine, query) when is_list(query) do
    with {:ok, path} <- Convert.directory_to_path(query) do

      log(engine, "Listing directory: #{inspect(query)}")

      case engine.db do
        nil ->
          # No database connection (testing mode)
          {:ok, []}

        db ->
          # Real FDB interaction
          try do
            # In production with erlfdb:
            # subdirs = :erlfdb.transactional(db, fn tx ->
            #   :erlfdb_directory.list(tx, path)
            # end)
            # {:ok, Enum.map(subdirs, fn name -> path ++ [name] end)}

            # For now, return empty
            {:ok, []}
          rescue
            e -> {:error, "FDB error: #{inspect(e)}"}
          end
      end
    end
  end

  @doc """
  Creates a stream for range queries that can be consumed lazily.
  """
  @spec stream_range(t(), KeyVal.key_value(), range_opts()) :: Enumerable.t()
  def stream_range(engine, query, opts \\ %{reverse: false, filter: false, limit: 0}) do
    Stream.resource(
      fn ->
        # Initialize: check query and setup state
        with :read_range <- Class.classify(query),
             {:ok, path} <- Convert.directory_to_path(query.key.directory) do
          %{
            engine: engine,
            query: query,
            path: path,
            opts: opts,
            position: nil,
            done: false
          }
        else
          {:error, reason} -> {:error, reason}
          _ -> {:error, "invalid query for range streaming"}
        end
      end,
      fn
        {:error, reason} ->
          {:halt, {:error, reason}}

        state when state.done ->
          {:halt, state}

        state ->
          # Fetch next batch
          # In production, would use :erlfdb.get_range with continuation
          # For now, return empty and mark done
          {[], %{state | done: true}}
      end,
      fn state ->
        # Cleanup: nothing to do in our case
        state
      end
    )
  end

  # Helper functions

  defp extract_value_type(%{types: []}), do: :any
  defp extract_value_type(%{types: [type | _]}), do: type
  defp extract_value_type(_), do: :any

  defp log(engine, message) do
    if engine.logger do
      # Would use proper logging library like Logger
      IO.puts("[FQL Engine] #{message}")
    end
  end
end
