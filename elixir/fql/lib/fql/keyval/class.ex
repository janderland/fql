defmodule Fql.KeyVal.Class do
  @moduledoc """
  Classifies a key-value by the kind of operation it represents.

  ## Classes

  - `:constant` - No variables, used for set or returned by get operations
  - `:vstamp_key` - Constant with VStampFuture in the key (set only)
  - `:vstamp_val` - Constant with VStampFuture in the value (set only)
  - `:clear` - Clear operation
  - `:read_single` - Single key-value read operation
  - `:read_range` - Multiple key-value read operation
  """

  alias Fql.KeyVal

  @type class ::
    :constant
    | :vstamp_key
    | :vstamp_val
    | :clear
    | :read_single
    | :read_range
    | {:invalid, String.t()}

  @doc """
  Classifies the given KeyValue.
  """
  @spec classify(KeyVal.key_value()) :: class()
  def classify(kv) do
    dir_attr = get_attributes_of_dir(kv.key.directory)
    key_attr = merge_attributes(dir_attr, get_attributes_of_tup(kv.key.tuple))
    kv_attr = merge_attributes(key_attr, get_attributes_of_val(kv.value))

    cond do
      # KeyValues should never contain nil
      kv_attr.has_nil ->
        invalid_class(kv_attr)

      # KeyValues should contain at most 1 VStampFuture
      kv_attr.vstamp_futures > 1 ->
        invalid_class(kv_attr)

      # Ensure at most one of: vstamp_futures, has_variable, has_clear
      count_conditions(kv_attr) > 1 ->
        invalid_class(kv_attr)

      # Classification based on attributes
      key_attr.has_variable ->
        :read_range

      kv_attr.has_variable ->
        :read_single

      kv_attr.vstamp_futures > 0 ->
        if key_attr.vstamp_futures > 0 do
          :vstamp_key
        else
          :vstamp_val
        end

      kv_attr.has_clear ->
        :clear

      true ->
        :constant
    end
  end

  # Attributes structure
  defp merge_attributes(attr1, attr2) do
    %{
      vstamp_futures: attr1.vstamp_futures + attr2.vstamp_futures,
      has_variable: attr1.has_variable or attr2.has_variable,
      has_clear: attr1.has_clear or attr2.has_clear,
      has_nil: attr1.has_nil or attr2.has_nil
    }
  end

  defp count_conditions(attr) do
    [
      attr.vstamp_futures > 0,
      attr.has_variable,
      attr.has_clear
    ]
    |> Enum.count(&(&1))
  end

  defp invalid_class(attr) do
    parts =
      []
      |> maybe_add(attr.vstamp_futures > 0, "vstamps:#{attr.vstamp_futures}")
      |> maybe_add(attr.has_variable, "var")
      |> maybe_add(attr.has_clear, "clear")
      |> maybe_add(attr.has_nil, "nil")
      |> Enum.join(",")

    {:invalid, "[#{parts}]"}
  end

  defp maybe_add(list, true, item), do: [item | list]
  defp maybe_add(list, false, _item), do: list

  # Get attributes from directory
  defp get_attributes_of_dir(dir) do
    Enum.reduce(dir, empty_attributes(), fn elem, acc ->
      merge_attributes(acc, get_attributes_of_dir_elem(elem))
    end)
  end

  defp get_attributes_of_dir_elem(%{types: _}), do: %{empty_attributes() | has_variable: true}
  defp get_attributes_of_dir_elem(_), do: empty_attributes()

  # Get attributes from tuple
  defp get_attributes_of_tup(tup) do
    Enum.reduce(tup, empty_attributes(), fn elem, acc ->
      merge_attributes(acc, get_attributes_of_tup_elem(elem))
    end)
  end

  defp get_attributes_of_tup_elem(elem) when is_list(elem) do
    get_attributes_of_tup(elem)
  end

  defp get_attributes_of_tup_elem(:nil) do
    %{empty_attributes() | has_nil: true}
  end

  defp get_attributes_of_tup_elem(%{types: _}) do
    %{empty_attributes() | has_variable: true}
  end

  defp get_attributes_of_tup_elem(:maybe_more) do
    empty_attributes()
  end

  defp get_attributes_of_tup_elem(%{tx_version: _, user_version: _}) do
    # VStamp
    empty_attributes()
  end

  defp get_attributes_of_tup_elem(%{user_version: _}) do
    # VStampFuture
    %{empty_attributes() | vstamp_futures: 1}
  end

  defp get_attributes_of_tup_elem(_), do: empty_attributes()

  # Get attributes from value
  defp get_attributes_of_val(val) when is_list(val) do
    get_attributes_of_tup(val)
  end

  defp get_attributes_of_val(:nil) do
    %{empty_attributes() | has_nil: true}
  end

  defp get_attributes_of_val(%{types: _}) do
    %{empty_attributes() | has_variable: true}
  end

  defp get_attributes_of_val(:clear) do
    %{empty_attributes() | has_clear: true}
  end

  defp get_attributes_of_val(%{tx_version: _, user_version: _}) do
    # VStamp
    empty_attributes()
  end

  defp get_attributes_of_val(%{user_version: _}) do
    # VStampFuture
    %{empty_attributes() | vstamp_futures: 1}
  end

  defp get_attributes_of_val(_), do: empty_attributes()

  defp empty_attributes do
    %{
      vstamp_futures: 0,
      has_variable: false,
      has_clear: false,
      has_nil: false
    }
  end
end
