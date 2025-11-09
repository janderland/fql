defmodule Fql.Parser.Scanner do
  @moduledoc """
  Tokenizes FQL query strings.
  """

  @type token_kind ::
    :whitespace
    | :newline
    | :escape
    | :other
    | :end
    | :key_val_sep
    | :dir_sep
    | :tup_start
    | :tup_end
    | :tup_sep
    | :var_start
    | :var_end
    | :var_sep
    | :str_mark
    | :stamp_start
    | :stamp_sep
    | :reserved

  @type token :: %{
    kind: token_kind(),
    value: String.t()
  }

  # Special characters
  @key_val_sep "="
  @dir_sep "/"
  @tup_start "("
  @tup_end ")"
  @tup_sep ","
  @var_start "<"
  @var_end ">"
  @var_sep "|"
  @str_mark "\""
  @stamp_start "@"
  @stamp_sep ":"
  @escape "\\"

  @whitespace [?\s, ?\t]
  @newline [?\n, ?\r]

  @doc """
  Scans a query string and returns a list of tokens.
  """
  @spec scan(String.t()) :: {:ok, [token()]} | {:error, String.t()}
  def scan(query) do
    scan_tokens(query, [])
  end

  defp scan_tokens("", acc), do: {:ok, Enum.reverse(acc)}

  defp scan_tokens(query, acc) do
    case scan_next_token(query) do
      {:ok, token, rest} ->
        scan_tokens(rest, [token | acc])
      {:error, reason} ->
        {:error, reason}
    end
  end

  defp scan_next_token(query) do
    cond do
      String.starts_with?(query, @escape) ->
        scan_escape(query)

      String.starts_with?(query, @key_val_sep) ->
        {:ok, token(:key_val_sep, @key_val_sep), String.slice(query, 1..-1)}

      String.starts_with?(query, @dir_sep) ->
        {:ok, token(:dir_sep, @dir_sep), String.slice(query, 1..-1)}

      String.starts_with?(query, @tup_start) ->
        {:ok, token(:tup_start, @tup_start), String.slice(query, 1..-1)}

      String.starts_with?(query, @tup_end) ->
        {:ok, token(:tup_end, @tup_end), String.slice(query, 1..-1)}

      String.starts_with?(query, @tup_sep) ->
        {:ok, token(:tup_sep, @tup_sep), String.slice(query, 1..-1)}

      String.starts_with?(query, @var_start) ->
        {:ok, token(:var_start, @var_start), String.slice(query, 1..-1)}

      String.starts_with?(query, @var_end) ->
        {:ok, token(:var_end, @var_end), String.slice(query, 1..-1)}

      String.starts_with?(query, @var_sep) ->
        {:ok, token(:var_sep, @var_sep), String.slice(query, 1..-1)}

      String.starts_with?(query, @str_mark) ->
        {:ok, token(:str_mark, @str_mark), String.slice(query, 1..-1)}

      String.starts_with?(query, @stamp_start) ->
        {:ok, token(:stamp_start, @stamp_start), String.slice(query, 1..-1)}

      String.starts_with?(query, @stamp_sep) ->
        {:ok, token(:stamp_sep, @stamp_sep), String.slice(query, 1..-1)}

      true ->
        scan_other(query)
    end
  end

  defp scan_escape(query) do
    case String.slice(query, 0..1) do
      <<@escape::utf8, next::utf8>> when next != 0 ->
        {:ok, token(:escape, <<next::utf8>>), String.slice(query, 2..-1)}
      _ ->
        {:error, "invalid escape sequence"}
    end
  end

  defp scan_other(query) do
    first_char = String.first(query)

    cond do
      first_char in @whitespace ->
        scan_whitespace(query)

      first_char in @newline ->
        scan_newline(query)

      true ->
        scan_word(query)
    end
  end

  defp scan_whitespace(query) do
    {ws, rest} = take_while(query, fn c -> c in @whitespace end)
    {:ok, token(:whitespace, ws), rest}
  end

  defp scan_newline(query) do
    {nl, rest} = take_while(query, fn c -> c in @whitespace or c in @newline end)
    {:ok, token(:newline, nl), rest}
  end

  defp scan_word(query) do
    special_chars = [
      @key_val_sep, @dir_sep, @tup_start, @tup_end, @tup_sep,
      @var_start, @var_end, @var_sep, @str_mark, @stamp_start,
      @stamp_sep, @escape
    ]

    {word, rest} = take_while(query, fn c ->
      not String.starts_with?(c, special_chars) and
      c not in @whitespace and
      c not in @newline
    end)

    {:ok, token(:other, word), rest}
  end

  defp take_while(string, fun) do
    take_while(string, "", fun)
  end

  defp take_while("", acc, _fun), do: {acc, ""}

  defp take_while(string, acc, fun) do
    char = String.first(string)
    if fun.(char) do
      take_while(String.slice(string, 1..-1), acc <> char, fun)
    else
      {acc, string}
    end
  end

  defp token(kind, value) do
    %{kind: kind, value: value}
  end
end
