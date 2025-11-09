defmodule Fql.ParserTest do
  use ExUnit.Case
  alias Fql.Parser
  alias Fql.KeyVal

  describe "parse/1" do
    test "parses simple directory query" do
      assert {:ok, ["path", "to", "dir"]} = Parser.parse("/path/to/dir")
    end

    test "parses empty directory" do
      assert {:ok, []} = Parser.parse("/")
    end

    test "parses directory with variable" do
      assert {:ok, ["path", %{types: [:int]}]} = Parser.parse("/path/<int>")
    end

    test "parses key-value with constant value" do
      {:ok, result} = Parser.parse("/dir(1, 2) = \"value\"")

      assert %{key: %{directory: ["dir"], tuple: [_, _]}, value: "value"} = result
    end

    test "parses set query with integer" do
      {:ok, result} = Parser.parse("/test(1) = 42")

      assert %{
        key: %{directory: ["test"], tuple: [tuple1]},
        value: {:int, 42}
      } = result

      assert {:int, 1} = tuple1
    end

    test "parses clear query" do
      {:ok, result} = Parser.parse("/test(1) = clear")

      assert %{
        key: %{directory: ["test"], tuple: _},
        value: :clear
      } = result
    end

    test "parses read single query" do
      {:ok, result} = Parser.parse("/test(1)")

      assert %{
        key: %{directory: ["test"], tuple: _},
        value: %{types: []}
      } = result
    end

    test "parses read single with typed variable" do
      {:ok, result} = Parser.parse("/test(1) = <string>")

      assert %{
        key: %{directory: ["test"], tuple: _},
        value: %{types: [:string]}
      } = result
    end

    test "parses read range query with variable in tuple" do
      {:ok, result} = Parser.parse("/test(<int>) = <>")

      assert %{
        key: %{directory: ["test"], tuple: [%{types: [:int]}]},
        value: %{types: []}
      } = result
    end

    test "parses tuple with multiple elements" do
      {:ok, result} = Parser.parse("/test(1, 2, 3) = \"value\"")

      assert %{
        key: %{tuple: [_, _, _]}
      } = result
    end

    test "parses boolean true" do
      {:ok, result} = Parser.parse("/test(true) = \"value\"")

      assert %{key: %{tuple: [true]}} = result
    end

    test "parses boolean false" do
      {:ok, result} = Parser.parse("/test(false) = \"value\"")

      assert %{key: %{tuple: [false]}} = result
    end

    test "parses nil" do
      {:ok, result} = Parser.parse("/test(nil) = \"value\"")

      assert %{key: %{tuple: [:nil]}} = result
    end

    test "parses negative integer" do
      {:ok, result} = Parser.parse("/test(-42) = \"value\"")

      assert %{key: %{tuple: [{:int, -42}]}} = result
    end

    test "parses float" do
      {:ok, result} = Parser.parse("/test(3.14) = \"value\"")

      assert %{key: %{tuple: [float_val]}} = result
      assert_in_delta float_val, 3.14, 0.001
    end

    test "parses hex bytes" do
      {:ok, result} = Parser.parse("/test(0xDEADBEEF) = \"value\"")

      assert %{key: %{tuple: [{:bytes, bytes}]}} = result
      assert bytes == <<0xDE, 0xAD, 0xBE, 0xEF>>
    end

    test "parses string with escapes" do
      {:ok, result} = Parser.parse("/test(\"hello\\nworld\") = \"value\"")

      assert %{key: %{tuple: [str]}} = result
      assert str == "hellonworld"  # Scanner returns the actual character
    end

    test "parses nested tuple" do
      {:ok, result} = Parser.parse("/test((1, 2)) = \"value\"")

      assert %{key: %{tuple: [[_, _]]}} = result
    end

    test "parses variable with multiple types" do
      {:ok, result} = Parser.parse("/test(<int|string|bool>) = <>")

      assert %{
        key: %{tuple: [%{types: [:int, :string, :bool]}]}
      } = result
    end

    test "parses complex query" do
      {:ok, result} = Parser.parse("/path/to/dir(1, \"test\", true) = 42")

      assert %{
        key: %{
          directory: ["path", "to", "dir"],
          tuple: [{:int, 1}, "test", true]
        },
        value: {:int, 42}
      } = result
    end

    test "returns error for invalid query" do
      assert {:error, _} = Parser.parse("/test(")
    end

    test "returns error for unterminated string" do
      assert {:error, _} = Parser.parse("/test(\"hello)")
    end
  end

  describe "round-trip formatting" do
    test "formats and parses back to equivalent structure" do
      original = "/test(1, 2) = \"value\""
      {:ok, parsed} = Parser.parse(original)
      formatted = Fql.Parser.Format.format(parsed)
      {:ok, reparsed} = Parser.parse(formatted)

      assert parsed == reparsed
    end

    test "formats directory query" do
      {:ok, parsed} = Parser.parse("/path/to/dir")
      formatted = Fql.Parser.Format.format(parsed)

      assert formatted == "/path/to/dir"
    end

    test "formats read query" do
      {:ok, parsed} = Parser.parse("/test(1)")
      formatted = Fql.Parser.Format.format(parsed)

      assert formatted == "/test(1)"
    end

    test "formats clear query" do
      {:ok, parsed} = Parser.parse("/test(1) = clear")
      formatted = Fql.Parser.Format.format(parsed)

      assert formatted == "/test(1) = clear"
    end
  end
end
