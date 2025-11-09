defmodule Fql.KeyVal.Convert do
  @moduledoc """
  Conversion utilities between FQL types and FoundationDB types.
  """

  alias Fql.KeyVal
  alias Fql.KeyVal.Tuple

  @doc """
  Converts a Directory to a list of strings.
  Returns error if directory contains variables.
  """
  @spec directory_to_path(KeyVal.directory()) :: {:ok, [String.t()]} | {:error, String.t()}
  def directory_to_path(directory) when is_list(directory) do
    convert_dir_elements(directory, [])
  end

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

  @doc """
  Creates a directory from a list of strings.
  """
  @spec path_to_directory([String.t()]) :: KeyVal.directory()
  def path_to_directory(path) when is_list(path) do
    path
  end

  @doc """
  Converts a KeyVal.Tuple to FDB tuple format.
  """
  @spec to_fdb_tuple(KeyVal.tuple()) :: {:ok, list()} | {:error, String.t()}
  def to_fdb_tuple(tuple) do
    Tuple.to_fdb_tuple(tuple)
  end

  @doc """
  Converts from FDB tuple format to KeyVal.Tuple.
  """
  @spec from_fdb_tuple(list()) :: KeyVal.tuple()
  def from_fdb_tuple(fdb_tuple) do
    Tuple.from_fdb_tuple(fdb_tuple)
  end

  @doc """
  Checks if a tuple has a versionstamp future in it.
  """
  @spec has_versionstamp_future?(KeyVal.tuple()) :: boolean()
  def has_versionstamp_future?(tuple) when is_list(tuple) do
    Enum.any?(tuple, fn
      %{user_version: _} = elem when not is_map_key(elem, :tx_version) -> true
      nested when is_list(nested) -> has_versionstamp_future?(nested)
      _ -> false
    end)
  end

  @doc """
  Extracts the versionstamp position in a tuple (if any).
  Returns the index of the versionstamp or nil.
  """
  @spec versionstamp_position(KeyVal.tuple()) :: integer() | nil
  def versionstamp_position(tuple) when is_list(tuple) do
    tuple
    |> Enum.with_index()
    |> Enum.find_value(fn {elem, idx} ->
      case elem do
        %{user_version: _} when not is_map_key(elem, :tx_version) -> idx
        _ -> nil
      end
    end)
  end
end
