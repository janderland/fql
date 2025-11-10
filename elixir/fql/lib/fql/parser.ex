defmodule Fql.Parser do
  @moduledoc """
  Converts FQL query strings into KeyVal structures.

  The parser uses a state machine to process tokens from the scanner
  and build up the query structure.
  """

  alias Fql.KeyVal
  alias Fql.Parser.Scanner

  @type parse_error :: {:error, String.t()}

  @doc """
  Parses a query string into a KeyVal query structure.
  """
  @spec parse(String.t()) :: {:ok, KeyVal.query()} | parse_error()
  def parse(query) do
    with {:ok, tokens} <- Scanner.scan(query),
         {:ok, result} <- parse_tokens(tokens) do
      {:ok, result}
    end
  end

  defp parse_tokens(tokens) do
    # Remove whitespace tokens
    tokens = Enum.reject(tokens, fn t -> t.kind in [:whitespace, :newline] end)

    state = %{
      tokens: tokens,
      pos: 0,
      directory: [],
      tuple: [],
      value: nil,
      current_string: "",
      current_var_types: []
    }

    parse_query(state)
  end

  defp parse_query(state) do
    cond do
      # Check if it's a directory query (starts with /)
      peek_token(state, :dir_sep) ->
        parse_directory_query(state)

      # Otherwise it's a key-value query
      true ->
        parse_key_value_query(state)
    end
  end

  defp parse_directory_query(state) do
    case parse_directory(state) do
      {:ok, directory, state} ->
        if state.pos >= length(state.tokens) do
          {:ok, directory}
        else
          {:error, "unexpected tokens after directory"}
        end

      {:error, reason} ->
        {:error, reason}
    end
  end

  defp parse_key_value_query(state) do
    with {:ok, directory, state} <- parse_directory(state),
         {:ok, tuple, state} <- parse_tuple(state),
         state = %{state | directory: directory, tuple: tuple},
         {:ok, value, state} <- parse_value_part(state) do

      key = KeyVal.key(directory, tuple)
      kv = KeyVal.key_value(key, value)
      {:ok, kv}
    end
  end

  defp parse_directory(state) do
    parse_directory_elements(state, [])
  end

  defp parse_directory_elements(state, acc) do
    cond do
      peek_token(state, :dir_sep) ->
        state = advance(state)
        case current_token(state) do
          %{kind: :other, value: name} ->
            state = advance(state)
            parse_directory_elements(state, [name | acc])

          %{kind: :var_start} ->
            case parse_variable(state) do
              {:ok, var, state} ->
                parse_directory_elements(state, [var | acc])
              {:error, reason} ->
                {:error, reason}
            end

          _ ->
            # End of directory
            {:ok, Enum.reverse(acc), state}
        end

      true ->
        {:ok, Enum.reverse(acc), state}
    end
  end

  defp parse_tuple(state) do
    if peek_token(state, :tup_start) do
      state = advance(state)  # consume (
      parse_tuple_elements(state, [])
    else
      {:ok, [], state}
    end
  end

  defp parse_tuple_elements(state, acc) do
    cond do
      peek_token(state, :tup_end) ->
        state = advance(state)  # consume )
        {:ok, Enum.reverse(acc), state}

      peek_token(state, :tup_sep) ->
        state = advance(state)  # consume ,
        parse_tuple_elements(state, acc)

      true ->
        case parse_tuple_element(state) do
          {:ok, elem, state} ->
            parse_tuple_elements(state, [elem | acc])
          {:error, reason} ->
            {:error, reason}
        end
    end
  end

  defp parse_tuple_element(state) do
    case current_token(state) do
      %{kind: :other, value: "nil"} ->
        {:ok, :nil, advance(state)}

      %{kind: :other, value: "true"} ->
        {:ok, true, advance(state)}

      %{kind: :other, value: "false"} ->
        {:ok, false, advance(state)}

      %{kind: :other, value: value} ->
        parse_number_or_string(value, advance(state))

      %{kind: :str_mark} ->
        parse_string(advance(state))

      %{kind: :var_start} ->
        parse_variable(state)

      %{kind: :tup_start} ->
        parse_tuple(state)

      _ ->
        {:error, "unexpected token in tuple"}
    end
  end

  defp parse_value_part(state) do
    cond do
      peek_token(state, :key_val_sep) ->
        state = advance(state)  # consume =
        parse_value(state)

      true ->
        # No value means it's a read query
        {:ok, KeyVal.variable([]), state}
    end
  end

  defp parse_value(state) do
    case current_token(state) do
      %{kind: :other, value: "clear"} ->
        {:ok, :clear, advance(state)}

      %{kind: :var_start} ->
        parse_variable(state)

      _ ->
        parse_tuple_element(state)
    end
  end

  defp parse_variable(state) do
    if peek_token(state, :var_start) do
      state = advance(state)  # consume <
      parse_variable_types(state, [])
    else
      {:error, "expected variable start"}
    end
  end

  defp parse_variable_types(state, acc) do
    cond do
      peek_token(state, :var_end) ->
        state = advance(state)  # consume >
        {:ok, KeyVal.variable(Enum.reverse(acc)), state}

      peek_token(state, :var_sep) ->
        state = advance(state)  # consume |
        parse_variable_types(state, acc)

      true ->
        case current_token(state) do
          %{kind: :other, value: type} ->
            type_atom = parse_type(type)
            state = advance(state)
            parse_variable_types(state, [type_atom | acc])

          _ ->
            {:error, "unexpected token in variable"}
        end
    end
  end

  defp parse_type("int"), do: :int
  defp parse_type("uint"), do: :uint
  defp parse_type("bool"), do: :bool
  defp parse_type("float"), do: :float
  defp parse_type("string"), do: :string
  defp parse_type("bytes"), do: :bytes
  defp parse_type("uuid"), do: :uuid
  defp parse_type("tuple"), do: :tuple
  defp parse_type("vstamp"), do: :vstamp
  defp parse_type(_), do: :any

  defp parse_number_or_string(value, state) do
    cond do
      String.match?(value, ~r/^-?\d+$/) ->
        {:ok, KeyVal.int(String.to_integer(value)), state}

      String.match?(value, ~r/^\d+\.\d+$/) ->
        {:ok, String.to_float(value), state}

      String.starts_with?(value, "0x") ->
        case parse_hex(value) do
          {:ok, bytes} -> {:ok, KeyVal.bytes(bytes), state}
          {:error, reason} -> {:error, reason}
        end

      true ->
        {:ok, value, state}
    end
  end

  defp parse_hex(value) do
    hex_string = String.slice(value, 2..-1)
    case Base.decode16(hex_string, case: :mixed) do
      {:ok, bytes} -> {:ok, bytes}
      :error -> {:error, "invalid hex string"}
    end
  end

  defp parse_string(state) do
    collect_string_contents(state, "")
  end

  defp collect_string_contents(state, acc) do
    case current_token(state) do
      %{kind: :str_mark} ->
        # End of string
        {:ok, acc, advance(state)}

      %{kind: :escape, value: char} ->
        # Escaped character - add the actual character
        collect_string_contents(advance(state), acc <> char)

      %{kind: _, value: value} ->
        # Regular content
        collect_string_contents(advance(state), acc <> value)

      nil ->
        {:error, "unterminated string"}
    end
  end

  defp collect_until(state, target_kind, acc) do
    case current_token(state) do
      %{kind: ^target_kind} ->
        {:ok, acc, advance(state)}

      %{kind: _, value: value} ->
        collect_until(advance(state), target_kind, acc <> value)

      nil ->
        {:error, "unexpected end of input"}
    end
  end

  # Token helpers
  defp current_token(state) do
    if state.pos < length(state.tokens) do
      Enum.at(state.tokens, state.pos)
    else
      nil
    end
  end

  defp peek_token(state, kind) do
    case current_token(state) do
      %{kind: ^kind} -> true
      _ -> false
    end
  end

  defp advance(state) do
    %{state | pos: state.pos + 1}
  end
end
