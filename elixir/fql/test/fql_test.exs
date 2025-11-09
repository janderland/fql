defmodule FqlTest do
  use ExUnit.Case
  doctest Fql

  test "version returns a string" do
    assert is_binary(Fql.version())
  end
end
