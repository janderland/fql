defmodule Fql.KeyVal.TupleTest do
  use ExUnit.Case
  alias Fql.KeyVal
  alias Fql.KeyVal.Tuple

  describe "to_fdb_tuple/1" do
    test "converts empty tuple" do
      assert {:ok, []} = Tuple.to_fdb_tuple([])
    end

    test "converts simple tuple" do
      tuple = [KeyVal.int(1), KeyVal.int(2), KeyVal.int(3)]
      assert {:ok, [1, 2, 3]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts mixed types" do
      tuple = [
        KeyVal.int(42),
        KeyVal.uint(100),
        true,
        3.14,
        "hello"
      ]

      assert {:ok, [42, 100, true, 3.14, "hello"]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts nil" do
      tuple = [:nil]
      assert {:ok, [nil]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts bytes" do
      bytes = <<1, 2, 3>>
      tuple = [KeyVal.bytes(bytes)]
      assert {:ok, [^bytes]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts uuid" do
      uuid = <<0::128>>
      tuple = [KeyVal.uuid(uuid)]
      assert {:ok, [{:uuid, ^uuid}]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts vstamp" do
      vstamp = KeyVal.vstamp(<<0::80>>, 123)
      tuple = [vstamp]
      assert {:ok, [{:versionstamp, _, 123}]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts vstamp_future" do
      vstamp_future = KeyVal.vstamp_future(456)
      tuple = [vstamp_future]
      assert {:ok, [{:versionstamp_future, 456}]} = Tuple.to_fdb_tuple(tuple)
    end

    test "converts nested tuple" do
      tuple = [[KeyVal.int(1), KeyVal.int(2)], KeyVal.int(3)]
      assert {:ok, [[1, 2], 3]} = Tuple.to_fdb_tuple(tuple)
    end

    test "rejects maybe_more" do
      tuple = [:maybe_more]
      assert {:error, _} = Tuple.to_fdb_tuple(tuple)
    end

    test "rejects variable" do
      tuple = [KeyVal.variable([:int])]
      assert {:error, _} = Tuple.to_fdb_tuple(tuple)
    end
  end

  describe "from_fdb_tuple/1" do
    test "converts empty tuple" do
      assert [] = Tuple.from_fdb_tuple([])
    end

    test "converts simple tuple" do
      fdb_tuple = [1, 2, 3]
      result = Tuple.from_fdb_tuple(fdb_tuple)

      assert [
        {:int, 1},
        {:int, 2},
        {:int, 3}
      ] = result
    end

    test "converts mixed types" do
      fdb_tuple = [42, true, 3.14, "hello"]
      result = Tuple.from_fdb_tuple(fdb_tuple)

      assert [
        {:int, 42},
        true,
        3.14,
        "hello"
      ] = result
    end

    test "converts nil" do
      fdb_tuple = [nil]
      assert [:nil] = Tuple.from_fdb_tuple(fdb_tuple)
    end

    test "converts uuid" do
      uuid = <<0::128>>
      fdb_tuple = [{:uuid, uuid}]
      result = Tuple.from_fdb_tuple(fdb_tuple)

      assert [{:uuid, ^uuid}] = result
    end

    test "converts vstamp" do
      tx_ver = <<0::80>>
      fdb_tuple = [{:versionstamp, tx_ver, 123}]
      result = Tuple.from_fdb_tuple(fdb_tuple)

      assert [%{tx_version: ^tx_ver, user_version: 123}] = result
    end

    test "converts nested tuple" do
      fdb_tuple = [[1, 2], 3]
      result = Tuple.from_fdb_tuple(fdb_tuple)

      assert [[{:int, 1}, {:int, 2}], {:int, 3}] = result
    end
  end

  describe "pack and unpack" do
    test "round-trips simple tuple" do
      tuple = [KeyVal.int(1), KeyVal.int(2), KeyVal.int(3)]
      {:ok, packed} = Tuple.pack(tuple)
      {:ok, unpacked} = Tuple.unpack(packed)

      assert unpacked == tuple
    end

    test "round-trips mixed types" do
      tuple = [
        KeyVal.int(42),
        true,
        3.14,
        "hello"
      ]

      {:ok, packed} = Tuple.pack(tuple)
      {:ok, unpacked} = Tuple.unpack(packed)

      assert unpacked == tuple
    end

    test "round-trips nested tuple" do
      tuple = [[KeyVal.int(1)], KeyVal.int(2)]
      {:ok, packed} = Tuple.pack(tuple)
      {:ok, unpacked} = Tuple.unpack(packed)

      assert unpacked == tuple
    end
  end

  describe "compare/2" do
    test "nil is less than everything" do
      assert Tuple.compare(:nil, KeyVal.int(1)) == :lt
      assert Tuple.compare(:nil, true) == :lt
      assert Tuple.compare(:nil, "hello") == :lt
    end

    test "nil equals nil" do
      assert Tuple.compare(:nil, :nil) == :eq
    end

    test "compares integers" do
      assert Tuple.compare(KeyVal.int(1), KeyVal.int(2)) == :lt
      assert Tuple.compare(KeyVal.int(2), KeyVal.int(1)) == :gt
      assert Tuple.compare(KeyVal.int(1), KeyVal.int(1)) == :eq
    end

    test "int is less than uint" do
      assert Tuple.compare(KeyVal.int(100), KeyVal.uint(50)) == :lt
    end

    test "compares strings" do
      assert Tuple.compare("a", "b") == :lt
      assert Tuple.compare("b", "a") == :gt
      assert Tuple.compare("a", "a") == :eq
    end

    test "compares booleans" do
      assert Tuple.compare(false, true) == :lt
      assert Tuple.compare(true, false) == :gt
      assert Tuple.compare(true, true) == :eq
    end
  end
end
