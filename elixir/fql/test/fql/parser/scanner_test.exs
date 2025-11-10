defmodule Fql.Parser.ScannerTest do
  use ExUnit.Case
  alias Fql.Parser.Scanner

  describe "scan/1" do
    test "scans directory separator" do
      assert {:ok, [%{kind: :dir_sep, value: "/"}]} = Scanner.scan("/")
    end

    test "scans key-value separator" do
      assert {:ok, [%{kind: :key_val_sep, value: "="}]} = Scanner.scan("=")
    end

    test "scans tuple delimiters" do
      assert {:ok, [
        %{kind: :tup_start, value: "("},
        %{kind: :tup_end, value: ")"}
      ]} = Scanner.scan("()")
    end

    test "scans variable delimiters" do
      assert {:ok, [
        %{kind: :var_start, value: "<"},
        %{kind: :var_end, value: ">"}
      ]} = Scanner.scan("<>")
    end

    test "scans other tokens" do
      assert {:ok, [%{kind: :other, value: "hello"}]} = Scanner.scan("hello")
    end

    test "scans whitespace" do
      assert {:ok, [
        %{kind: :other, value: "a"},
        %{kind: :whitespace, value: " "},
        %{kind: :other, value: "b"}
      ]} = Scanner.scan("a b")
    end

    test "scans complex query" do
      {:ok, tokens} = Scanner.scan("/dir(1, 2) = value")

      kinds = Enum.map(tokens, & &1.kind)
      assert :dir_sep in kinds
      assert :tup_start in kinds
      assert :tup_end in kinds
      assert :key_val_sep in kinds
    end

    test "scans escape sequences" do
      assert {:ok, [%{kind: :escape, value: "/"}]} = Scanner.scan("\\/")
    end
  end
end
