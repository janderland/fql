defmodule Fql.Parser.FormatTest do
  use ExUnit.Case
  alias Fql.Parser.Format
  alias Fql.KeyVal

  describe "format/1 with directories" do
    test "formats simple directory" do
      dir = ["path", "to", "dir"]
      assert Format.format(dir) == "/path/to/dir"
    end

    test "formats empty directory" do
      dir = []
      assert Format.format(dir) == ""
    end

    test "formats directory with variable" do
      dir = ["path", KeyVal.variable([:int])]
      assert Format.format(dir) == "/path/<int>"
    end
  end

  describe "format/1 with key-values" do
    test "formats set query" do
      kv = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        "value"
      )

      assert Format.format(kv) == "/test(1) = \"value\""
    end

    test "formats clear query" do
      kv = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        :clear
      )

      assert Format.format(kv) == "/test(1) = clear"
    end

    test "formats read single query" do
      kv = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        KeyVal.variable([])
      )

      assert Format.format(kv) == "/test(1)"
    end

    test "formats read with typed variable" do
      kv = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        KeyVal.variable([:string])
      )

      assert Format.format(kv) == "/test(1) = <string>"
    end
  end

  describe "format_tuple/1" do
    test "formats empty tuple" do
      assert Format.format_tuple([]) == ""
    end

    test "formats single element tuple" do
      assert Format.format_tuple([KeyVal.int(1)]) == "(1)"
    end

    test "formats multiple element tuple" do
      tuple = [KeyVal.int(1), KeyVal.int(2), KeyVal.int(3)]
      assert Format.format_tuple(tuple) == "(1, 2, 3)"
    end

    test "formats nil" do
      assert Format.format_tuple([:nil]) == "(nil)"
    end

    test "formats booleans" do
      assert Format.format_tuple([true]) == "(true)"
      assert Format.format_tuple([false]) == "(false)"
    end

    test "formats floats" do
      result = Format.format_tuple([3.14])
      assert String.contains?(result, "3.14")
    end

    test "formats strings" do
      assert Format.format_tuple(["hello"]) == "(\"hello\")"
    end

    test "formats bytes as hex" do
      result = Format.format_tuple([KeyVal.bytes(<<0xDE, 0xAD>>)])
      assert result == "(0xDEAD)"
    end

    test "formats uuid as hex" do
      uuid = <<1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16>>
      result = Format.format_tuple([KeyVal.uuid(uuid)])
      assert result == "(0x0102030405060708090A0B0C0D0E0F10)"
    end

    test "formats variables" do
      assert Format.format_tuple([KeyVal.variable([])]) == "(<>)"
      assert Format.format_tuple([KeyVal.variable([:int])]) == "(<int>)"
      assert Format.format_tuple([KeyVal.variable([:int, :string])]) == "(<int|string>)"
    end

    test "formats nested tuple" do
      tuple = [[KeyVal.int(1), KeyVal.int(2)], KeyVal.int(3)]
      assert Format.format_tuple(tuple) == "((1, 2), 3)"
    end
  end

  describe "format_variable/1" do
    test "formats empty variable" do
      assert Format.format_variable([]) == "<>"
    end

    test "formats single type variable" do
      assert Format.format_variable([:int]) == "<int>"
    end

    test "formats multi-type variable" do
      assert Format.format_variable([:int, :string, :bool]) == "<int|string|bool>"
    end

    test "formats all value types" do
      types = [:int, :uint, :bool, :float, :string, :bytes, :uuid, :tuple, :vstamp]
      result = Format.format_variable(types)
      assert result == "<int|uint|bool|float|string|bytes|uuid|tuple|vstamp>"
    end
  end

  describe "format_key/1" do
    test "formats key with directory and tuple" do
      key = KeyVal.key(["test"], [KeyVal.int(1), KeyVal.int(2)])
      assert Format.format_key(key) == "/test(1, 2)"
    end

    test "formats key with empty tuple" do
      key = KeyVal.key(["test"], [])
      assert Format.format_key(key) == "/test"
    end

    test "formats key with nested directory" do
      key = KeyVal.key(["path", "to", "test"], [KeyVal.int(1)])
      assert Format.format_key(key) == "/path/to/test(1)"
    end
  end

  describe "format_value/1" do
    test "formats clear" do
      assert Format.format_value(:clear) == "clear"
    end

    test "formats nil" do
      assert Format.format_value(:nil) == "nil"
    end

    test "formats variable" do
      assert Format.format_value(KeyVal.variable([:int])) == "<int>"
    end

    test "formats primitive values" do
      assert Format.format_value(KeyVal.int(42)) == "42"
      assert Format.format_value(true) == "true"
      assert Format.format_value("hello") == "\"hello\""
    end

    test "formats tuple value" do
      tuple = [KeyVal.int(1), KeyVal.int(2)]
      assert Format.format_value(tuple) == "(1, 2)"
    end
  end

  describe "round-trip" do
    test "formats and parses back for simple query" do
      original_str = "/test(1) = \"value\""
      {:ok, parsed} = Fql.Parser.parse(original_str)
      formatted = Format.format(parsed)
      {:ok, reparsed} = Fql.Parser.parse(formatted)

      assert parsed == reparsed
    end

    test "formats and parses back for complex query" do
      original_str = "/path/to/test(1, true, \"hello\") = 42"
      {:ok, parsed} = Fql.Parser.parse(original_str)
      formatted = Format.format(parsed)
      {:ok, reparsed} = Fql.Parser.parse(formatted)

      assert parsed == reparsed
    end

    test "formats and parses back for directory query" do
      original_str = "/path/to/dir"
      {:ok, parsed} = Fql.Parser.parse(original_str)
      formatted = Format.format(parsed)
      {:ok, reparsed} = Fql.Parser.parse(formatted)

      assert parsed == reparsed
    end

    test "formats and parses back for read query" do
      original_str = "/test(<int>) = <string>"
      {:ok, parsed} = Fql.Parser.parse(original_str)
      formatted = Format.format(parsed)
      {:ok, reparsed} = Fql.Parser.parse(formatted)

      assert parsed == reparsed
    end
  end
end
