defmodule Fql.Parser.Format do
  @moduledoc """
  Formats FQL queries back into strings.

  This module provides the inverse of the parser - it takes
  KeyVal structures and converts them back to FQL query strings.
  """

  alias Fql.KeyVal

  @doc """
  Formats a query into an FQL query string.
  """
  @spec format(KeyVal.query()) :: String.t()
  def format(query) when is_list(query) do
    # Directory query
    format_directory(query)
  end

  def format(%{key: key, value: value}) do
    # KeyValue query
    key_str = format_key(key)
    value_str = format_value(value)

    case value do
      %{types: []} -> key_str  # Read query with empty variable
      _ -> "#{key_str} = #{value_str}"
    end
  end

  @doc """
  Formats a key.
  """
  @spec format_key(KeyVal.key()) :: String.t()
  def format_key(%{directory: dir, tuple: tup}) do
    dir_str = format_directory(dir)
    tup_str = format_tuple(tup)

    if tup == [] do
      dir_str
    else
      "#{dir_str}#{tup_str}"
    end
  end

  @doc """
  Formats a directory.
  """
  @spec format_directory(KeyVal.directory()) :: String.t()
  def format_directory(dir) when is_list(dir) do
    Enum.map_join(dir, fn elem ->
      "/" <> format_dir_element(elem)
    end)
  end

  defp format_dir_element(elem) when is_binary(elem) do
    if needs_escape?(elem) do
      escape_string(elem)
    else
      elem
    end
  end

  defp format_dir_element(%{types: types}) do
    format_variable(types)
  end

  @doc """
  Formats a tuple.
  """
  @spec format_tuple(KeyVal.tuple()) :: String.t()
  def format_tuple([]), do: ""

  def format_tuple(tuple) when is_list(tuple) do
    elements = Enum.map_join(tuple, ", ", &format_tuple_element/1)
    "(#{elements})"
  end

  defp format_tuple_element(:nil), do: "nil"
  defp format_tuple_element(:maybe_more), do: "..."

  defp format_tuple_element(elem) when is_boolean(elem) do
    if elem, do: "true", else: "false"
  end

  defp format_tuple_element({:int, n}) when is_integer(n), do: Integer.to_string(n)
  defp format_tuple_element({:uint, n}) when is_integer(n), do: Integer.to_string(n)
  defp format_tuple_element(n) when is_integer(n), do: Integer.to_string(n)
  defp format_tuple_element(f) when is_float(f), do: Float.to_string(f)

  defp format_tuple_element(s) when is_binary(s) do
    if printable_string?(s) do
      "\"#{escape_string(s)}\""
    else
      {:bytes, s} |> format_tuple_element()
    end
  end

  defp format_tuple_element({:bytes, b}) when is_binary(b) do
    "0x" <> Base.encode16(b, case: :upper)
  end

  defp format_tuple_element({:uuid, u}) when byte_size(u) == 16 do
    "0x" <> Base.encode16(u, case: :upper)
  end

  defp format_tuple_element(%{tx_version: _tx, user_version: user}) do
    "@vstamp:#{user}"
  end

  defp format_tuple_element(%{user_version: user}) do
    "@vstamp_future:#{user}"
  end

  defp format_tuple_element(%{types: types}) do
    format_variable(types)
  end

  defp format_tuple_element(elem) when is_list(elem) do
    # Nested tuple
    format_tuple(elem)
  end

  defp format_tuple_element(_), do: "?"

  @doc """
  Formats a value.
  """
  @spec format_value(KeyVal.value()) :: String.t()
  def format_value(:clear), do: "clear"
  def format_value(:nil), do: "nil"

  def format_value(%{types: types}) do
    format_variable(types)
  end

  def format_value(val) when is_list(val) do
    # Tuple value
    format_tuple(val)
  end

  def format_value(val), do: format_tuple_element(val)

  @doc """
  Formats a variable with its type constraints.
  """
  @spec format_variable([KeyVal.value_type()]) :: String.t()
  def format_variable([]), do: "<>"

  def format_variable(types) do
    type_str = Enum.map_join(types, "|", &format_type/1)
    "<#{type_str}>"
  end

  defp format_type(:any), do: ""
  defp format_type(:int), do: "int"
  defp format_type(:uint), do: "uint"
  defp format_type(:bool), do: "bool"
  defp format_type(:float), do: "float"
  defp format_type(:string), do: "string"
  defp format_type(:bytes), do: "bytes"
  defp format_type(:uuid), do: "uuid"
  defp format_type(:tuple), do: "tuple"
  defp format_type(:vstamp), do: "vstamp"
  defp format_type(_), do: ""

  # Helper functions

  defp printable_string?(s) do
    String.valid?(s) and String.printable?(s)
  end

  defp needs_escape?(s) do
    String.contains?(s, ["\\", "\"", "/", "(", ")", "<", ">", "=", ",", "|"])
  end

  defp escape_string(s) do
    s
    |> String.replace("\\", "\\\\")
    |> String.replace("\"", "\\\"")
    |> String.replace("/", "\\/")
  end

  @doc """
  Pretty prints a query with indentation.
  """
  @spec pretty(KeyVal.query(), integer()) :: String.t()
  def pretty(query, indent \\ 0) do
    # For now, just use format
    # Could add multi-line formatting later
    String.duplicate("  ", indent) <> format(query)
  end
end
