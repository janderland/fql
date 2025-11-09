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
    # Would use :erlfdb.transact here
    fun.(engine)
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
    with {:ok, path} <- directory_to_path(query.key.directory),
         {:ok, value_bytes} <- Values.pack(query.value, engine.byte_order, class == :vstamp_val) do

      # Would interact with FDB here
      # :erlfdb.set(engine.db, key_bytes, value_bytes)
      log(engine, "Setting key-value: #{inspect(query)}")
      :ok
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
    with {:ok, _path} <- directory_to_path(query.key.directory) do
      # Would interact with FDB here
      # :erlfdb.clear(engine.db, key_bytes)
      log(engine, "Clearing key: #{inspect(query.key)}")
      :ok
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

  defp execute_read_single(engine, query, _opts) do
    with {:ok, _path} <- directory_to_path(query.key.directory) do
      # Would interact with FDB here
      # value_bytes = :erlfdb.get(engine.db, key_bytes)
      # unpack value_bytes based on variable types
      log(engine, "Reading single key-value: #{inspect(query.key)}")

      # Dummy response
      {:ok, query}
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

  defp execute_read_range(engine, query, _opts) do
    with {:ok, _path} <- directory_to_path(query.key.directory) do
      # Would interact with FDB here
      # range = :erlfdb.get_range(engine.db, start_key, end_key, opts)
      log(engine, "Reading range: #{inspect(query.key)}")

      # Dummy response
      {:ok, []}
    end
  end

  @doc """
  Lists directories matching the query schema.
  """
  @spec list_directory(t(), KeyVal.directory()) ::
    {:ok, [KeyVal.directory()]} | {:error, String.t()}
  def list_directory(engine, query) when is_list(query) do
    with {:ok, _path} <- directory_to_path(query) do
      # Would interact with FDB directory layer here
      # :erlfdb_directory.list(engine.db, path)
      log(engine, "Listing directory: #{inspect(query)}")

      # Dummy response
      {:ok, []}
    end
  end

  # Helper functions

  defp directory_to_path(directory) do
    path = Enum.map(directory, fn
      elem when is_binary(elem) -> elem
      %{types: _} -> {:error, "cannot convert variable to path"}
      _ -> {:error, "invalid directory element"}
    end)

    if Enum.any?(path, &match?({:error, _}, &1)) do
      {:error, "invalid directory"}
    else
      {:ok, path}
    end
  end

  defp log(engine, message) do
    if engine.logger do
      # Would use proper logging
      IO.puts("[FQL Engine] #{message}")
    end
  end
end
