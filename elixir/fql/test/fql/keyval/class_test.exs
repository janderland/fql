defmodule Fql.KeyVal.ClassTest do
  use ExUnit.Case
  alias Fql.KeyVal
  alias Fql.KeyVal.Class

  describe "classify/1" do
    test "classifies constant query" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.int(1)]),
        "value"
      )

      assert Class.classify(kv) == :constant
    end

    test "classifies clear query" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.int(1)]),
        :clear
      )

      assert Class.classify(kv) == :clear
    end

    test "classifies read_single query" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.int(1)]),
        KeyVal.variable([:string])
      )

      assert Class.classify(kv) == :read_single
    end

    test "classifies read_range query with variable in key" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.variable([:int])]),
        KeyVal.variable([])
      )

      assert Class.classify(kv) == :read_range
    end

    test "classifies vstamp_key query" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.vstamp_future(0)]),
        "value"
      )

      assert Class.classify(kv) == :vstamp_key
    end

    test "classifies vstamp_val query" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.int(1)]),
        KeyVal.vstamp_future(0)
      )

      assert Class.classify(kv) == :vstamp_val
    end

    test "classifies invalid query with nil" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [:nil]),
        "value"
      )

      assert {:invalid, _} = Class.classify(kv)
    end

    test "classifies invalid query with multiple vstamp futures" do
      kv = KeyVal.key_value(
        KeyVal.key(["dir"], [KeyVal.vstamp_future(0)]),
        KeyVal.vstamp_future(0)
      )

      assert {:invalid, _} = Class.classify(kv)
    end
  end
end
