defmodule Fql.EngineTest do
  use ExUnit.Case
  alias Fql.Engine
  alias Fql.KeyVal

  setup do
    # Create engine without FDB connection for testing
    engine = Engine.new(nil, byte_order: :big)
    {:ok, engine: engine}
  end

  describe "new/2" do
    test "creates engine with defaults" do
      engine = Engine.new(nil)
      assert engine.byte_order == :big
      assert engine.logger == nil
    end

    test "creates engine with options" do
      engine = Engine.new(nil, byte_order: :little, logger: true)
      assert engine.byte_order == :little
      assert engine.logger == true
    end
  end

  describe "set/2" do
    test "accepts constant query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        "value"
      )

      assert :ok = Engine.set(engine, query)
    end

    test "rejects invalid query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        KeyVal.variable([:string])
      )

      assert {:error, _} = Engine.set(engine, query)
    end

    test "rejects clear as set", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        :clear
      )

      assert {:error, _} = Engine.set(engine, query)
    end
  end

  describe "clear/2" do
    test "accepts clear query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        :clear
      )

      assert :ok = Engine.clear(engine, query)
    end

    test "rejects non-clear query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        "value"
      )

      assert {:error, _} = Engine.clear(engine, query)
    end
  end

  describe "read_single/3" do
    test "accepts read_single query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        KeyVal.variable([:string])
      )

      assert {:ok, _} = Engine.read_single(engine, query)
    end

    test "rejects range query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.variable([:int])]),
        KeyVal.variable([])
      )

      assert {:error, _} = Engine.read_single(engine, query)
    end
  end

  describe "read_range/3" do
    test "accepts read_range query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.variable([:int])]),
        KeyVal.variable([])
      )

      assert {:ok, []} = Engine.read_range(engine, query)
    end

    test "accepts read_range with options", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.variable([:int])]),
        KeyVal.variable([])
      )

      opts = %{reverse: true, filter: true, limit: 10}
      assert {:ok, []} = Engine.read_range(engine, query, opts)
    end

    test "rejects read_single query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.int(1)]),
        KeyVal.variable([])
      )

      assert {:error, _} = Engine.read_range(engine, query)
    end
  end

  describe "list_directory/2" do
    test "accepts directory query", %{engine: engine} do
      dir = ["test", "path"]
      assert {:ok, []} = Engine.list_directory(engine, dir)
    end

    test "rejects directory with variable", %{engine: engine} do
      dir = ["test", KeyVal.variable([])]
      assert {:error, _} = Engine.list_directory(engine, dir)
    end
  end

  describe "transact/2" do
    test "executes function in transaction", %{engine: engine} do
      result = Engine.transact(engine, fn tx_engine ->
        assert tx_engine.db == nil
        {:ok, :success}
      end)

      assert {:ok, :success} = result
    end

    test "handles transaction errors", %{engine: engine} do
      result = Engine.transact(engine, fn _tx_engine ->
        {:error, "something failed"}
      end)

      assert {:error, "something failed"} = result
    end
  end

  describe "stream_range/3" do
    test "creates stream for range query", %{engine: engine} do
      query = KeyVal.key_value(
        KeyVal.key(["test"], [KeyVal.variable([:int])]),
        KeyVal.variable([])
      )

      stream = Engine.stream_range(engine, query)
      assert is_function(stream)

      # Stream should be empty in test mode
      results = Enum.to_list(stream)
      assert results == []
    end
  end
end
