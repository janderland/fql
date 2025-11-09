defmodule Fql.KeyValTest do
  use ExUnit.Case
  alias Fql.KeyVal

  describe "all_types/0" do
    test "returns all value types" do
      types = KeyVal.all_types()
      assert :any in types
      assert :int in types
      assert :uint in types
      assert :bool in types
      assert :float in types
      assert :string in types
      assert :bytes in types
      assert :uuid in types
      assert :tuple in types
      assert :vstamp in types
    end
  end

  describe "key_value/2" do
    test "creates a key-value struct" do
      key = KeyVal.key([], [])
      value = "test"
      kv = KeyVal.key_value(key, value)

      assert kv.key == key
      assert kv.value == value
    end
  end

  describe "key/2" do
    test "creates a key struct" do
      directory = ["path", "to", "dir"]
      tuple = [KeyVal.int(1), KeyVal.int(2)]
      key = KeyVal.key(directory, tuple)

      assert key.directory == directory
      assert key.tuple == tuple
    end
  end

  describe "variable/1" do
    test "creates a variable with types" do
      var = KeyVal.variable([:int, :string])
      assert var.types == [:int, :string]
    end

    test "creates a variable with no types" do
      var = KeyVal.variable()
      assert var.types == []
    end
  end

  describe "primitive types" do
    test "int/1 creates tagged int" do
      assert KeyVal.int(42) == {:int, 42}
    end

    test "uint/1 creates tagged uint" do
      assert KeyVal.uint(42) == {:uint, 42}
    end

    test "uuid/1 creates tagged uuid" do
      uuid_bytes = <<0::128>>
      assert KeyVal.uuid(uuid_bytes) == {:uuid, uuid_bytes}
    end

    test "bytes/1 creates tagged bytes" do
      assert KeyVal.bytes("hello") == {:bytes, "hello"}
    end
  end

  describe "vstamp/2" do
    test "creates a vstamp struct" do
      tx_version = <<0::80>>
      user_version = 123
      vstamp = KeyVal.vstamp(tx_version, user_version)

      assert vstamp.tx_version == tx_version
      assert vstamp.user_version == user_version
    end
  end

  describe "vstamp_future/1" do
    test "creates a vstamp_future struct" do
      vstamp_future = KeyVal.vstamp_future(456)
      assert vstamp_future.user_version == 456
    end
  end
end
