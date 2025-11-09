defmodule Fql.KeyVal.Tuple do
  @moduledoc """
  Utilities for working with FDB tuples and converting between
  FQL tuples and FoundationDB tuple format.
  """

  alias Fql.KeyVal

  @doc """
  Converts a KeyVal.Tuple to an FDB tuple (list of erlang terms).
  """
  @spec to_fdb_tuple(KeyVal.tuple()) :: {:ok, list()} | {:error, String.t()}
  def to_fdb_tuple(tuple) when is_list(tuple) do
    convert_elements(tuple, [])
  end

  defp convert_elements([], acc) do
    {:ok, Enum.reverse(acc)}
  end

  defp convert_elements([elem | rest], acc) do
    case to_fdb_element(elem) do
      {:ok, fdb_elem} ->
        convert_elements(rest, [fdb_elem | acc])
      {:error, reason} ->
        {:error, reason}
    end
  end

  defp to_fdb_element(elem) when is_list(elem) do
    # Nested tuple
    to_fdb_tuple(elem)
  end

  defp to_fdb_element(:nil) do
    {:ok, nil}
  end

  defp to_fdb_element({:int, value}) when is_integer(value) do
    {:ok, value}
  end

  defp to_fdb_element({:uint, value}) when is_integer(value) and value >= 0 do
    {:ok, value}
  end

  defp to_fdb_element(value) when is_boolean(value) do
    {:ok, value}
  end

  defp to_fdb_element(value) when is_float(value) do
    {:ok, value}
  end

  defp to_fdb_element(value) when is_binary(value) do
    # Strings are binaries in Elixir
    {:ok, value}
  end

  defp to_fdb_element({:bytes, value}) when is_binary(value) do
    {:ok, value}
  end

  defp to_fdb_element({:uuid, value}) when byte_size(value) == 16 do
    {:ok, {:uuid, value}}
  end

  defp to_fdb_element(%{tx_version: tx_ver, user_version: user_ver}) do
    # VStamp - encode as special tuple element
    {:ok, {:versionstamp, tx_ver, user_ver}}
  end

  defp to_fdb_element(%{user_version: user_ver}) do
    # VStampFuture - encode as incomplete versionstamp
    {:ok, {:versionstamp_future, user_ver}}
  end

  defp to_fdb_element(:maybe_more) do
    {:error, "maybe_more cannot be converted to FDB element"}
  end

  defp to_fdb_element(%{types: _}) do
    {:error, "variables cannot be converted to FDB tuples"}
  end

  defp to_fdb_element(_) do
    {:error, "unknown tuple element type"}
  end

  @doc """
  Converts an FDB tuple (list of erlang terms) to a KeyVal.Tuple.
  """
  @spec from_fdb_tuple(list()) :: KeyVal.tuple()
  def from_fdb_tuple(fdb_tuple) when is_list(fdb_tuple) do
    Enum.map(fdb_tuple, &from_fdb_element/1)
  end

  defp from_fdb_element(nil), do: :nil

  defp from_fdb_element(value) when is_integer(value) do
    # Default to signed int
    KeyVal.int(value)
  end

  defp from_fdb_element(value) when is_boolean(value), do: value

  defp from_fdb_element(value) when is_float(value), do: value

  defp from_fdb_element(value) when is_binary(value) do
    # Check if it's a valid UTF-8 string or raw bytes
    case String.valid?(value) do
      true -> value
      false -> KeyVal.bytes(value)
    end
  end

  defp from_fdb_element({:uuid, value}) when byte_size(value) == 16 do
    KeyVal.uuid(value)
  end

  defp from_fdb_element({:versionstamp, tx_ver, user_ver}) do
    KeyVal.vstamp(tx_ver, user_ver)
  end

  defp from_fdb_element(value) when is_list(value) do
    # Nested tuple
    from_fdb_tuple(value)
  end

  defp from_fdb_element(_), do: :nil

  @doc """
  Packs a KeyVal.Tuple into bytes using FDB tuple encoding.
  This is a simplified implementation - in production, use :erlfdb.tuple.pack/1
  """
  @spec pack(KeyVal.tuple()) :: {:ok, binary()} | {:error, String.t()}
  def pack(tuple) do
    case to_fdb_tuple(tuple) do
      {:ok, fdb_tuple} ->
        # In production, use: :erlfdb_tuple.pack(fdb_tuple)
        # For now, return a simple encoding
        {:ok, :erlang.term_to_binary(fdb_tuple)}
      {:error, reason} ->
        {:error, reason}
    end
  end

  @doc """
  Unpacks bytes into a KeyVal.Tuple using FDB tuple encoding.
  This is a simplified implementation - in production, use :erlfdb.tuple.unpack/1
  """
  @spec unpack(binary()) :: {:ok, KeyVal.tuple()} | {:error, String.t()}
  def unpack(bytes) when is_binary(bytes) do
    try do
      # In production, use: :erlfdb_tuple.unpack(bytes)
      # For now, use erlang term format
      fdb_tuple = :erlang.binary_to_term(bytes)
      {:ok, from_fdb_tuple(fdb_tuple)}
    rescue
      _ -> {:error, "failed to unpack tuple"}
    end
  end

  @doc """
  Compares two tuple elements for ordering.
  Returns :lt, :eq, or :gt
  """
  @spec compare(any(), any()) :: :lt | :eq | :gt
  def compare(a, b) do
    cond do
      a == b -> :eq
      tuple_less_than(a, b) -> :lt
      true -> :gt
    end
  end

  defp tuple_less_than(:nil, _), do: true
  defp tuple_less_than(_, :nil), do: false

  defp tuple_less_than({:int, a}, {:int, b}), do: a < b
  defp tuple_less_than({:int, _}, {:uint, _}), do: true
  defp tuple_less_than({:uint, _}, {:int, _}), do: false
  defp tuple_less_than({:uint, a}, {:uint, b}), do: a < b

  defp tuple_less_than(a, b) when is_number(a) and is_number(b), do: a < b
  defp tuple_less_than(a, b) when is_binary(a) and is_binary(b), do: a < b
  defp tuple_less_than(a, b) when is_boolean(a) and is_boolean(b), do: not a and b

  defp tuple_less_than({:bytes, a}, {:bytes, b}), do: a < b
  defp tuple_less_than({:uuid, a}, {:uuid, b}), do: a < b

  defp tuple_less_than(_, _), do: false
end
