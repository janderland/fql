defmodule Fql.KeyVal.ValuesTest do
  use ExUnit.Case
  alias Fql.KeyVal
  alias Fql.KeyVal.Values

  describe "pack and unpack" do
    test "packs and unpacks bool" do
      assert {:ok, <<1>>} = Values.pack(true, :big, false)
      assert {:ok, <<0>>} = Values.pack(false, :big, false)

      assert {:ok, true} = Values.unpack(<<1>>, :bool, :big)
      assert {:ok, false} = Values.unpack(<<0>>, :bool, :big)
    end

    test "packs and unpacks int (big endian)" do
      assert {:ok, bytes} = Values.pack(KeyVal.int(42), :big, false)
      assert byte_size(bytes) == 8

      assert {:ok, {:int, 42}} = Values.unpack(bytes, :int, :big)
    end

    test "packs and unpacks int (little endian)" do
      assert {:ok, bytes} = Values.pack(KeyVal.int(42), :little, false)
      assert byte_size(bytes) == 8

      assert {:ok, {:int, 42}} = Values.unpack(bytes, :int, :little)
    end

    test "packs and unpacks uint" do
      assert {:ok, bytes} = Values.pack(KeyVal.uint(100), :big, false)
      assert byte_size(bytes) == 8

      assert {:ok, {:uint, 100}} = Values.unpack(bytes, :uint, :big)
    end

    test "packs and unpacks float" do
      assert {:ok, bytes} = Values.pack(3.14, :big, false)
      assert byte_size(bytes) == 8

      assert {:ok, float_val} = Values.unpack(bytes, :float, :big)
      assert_in_delta float_val, 3.14, 0.0001
    end

    test "packs and unpacks string" do
      assert {:ok, "hello"} = Values.pack("hello", :big, false)
      assert {:ok, "hello"} = Values.unpack("hello", :string, :big)
    end

    test "packs and unpacks bytes" do
      bytes = <<1, 2, 3, 4>>
      assert {:ok, ^bytes} = Values.pack(KeyVal.bytes(bytes), :big, false)
      assert {:ok, {:bytes, ^bytes}} = Values.unpack(bytes, :bytes, :big)
    end

    test "packs and unpacks uuid" do
      uuid = <<0::128>>
      assert {:ok, ^uuid} = Values.pack(KeyVal.uuid(uuid), :big, false)
      assert {:ok, {:uuid, ^uuid}} = Values.unpack(uuid, :uuid, :big)
    end

    test "packs and unpacks vstamp" do
      tx_ver = <<0::80>>
      vstamp = KeyVal.vstamp(tx_ver, 123)

      assert {:ok, bytes} = Values.pack(vstamp, :big, false)
      assert byte_size(bytes) == 12

      assert {:ok, unpacked} = Values.unpack(bytes, :vstamp, :big)
      assert unpacked.tx_version == tx_ver
      assert unpacked.user_version == 123
    end

    test "unpacks any type as bytes" do
      data = "arbitrary data"
      assert {:ok, {:bytes, ^data}} = Values.unpack(data, :any, :big)
    end

    test "returns error for nil value" do
      assert {:error, _} = Values.pack(nil, :big, false)
    end

    test "returns error for invalid bool size" do
      assert {:error, _} = Values.unpack(<<1, 2>>, :bool, :big)
    end

    test "returns error for invalid int size" do
      assert {:error, _} = Values.unpack(<<1, 2>>, :int, :big)
    end

    test "returns error for invalid uuid size" do
      assert {:error, _} = Values.unpack(<<1, 2>>, :uuid, :big)
    end
  end
end
