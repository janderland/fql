defmodule Fql.KeyVal.ConvertTest do
  use ExUnit.Case
  alias Fql.KeyVal
  alias Fql.KeyVal.Convert

  describe "directory_to_path/1" do
    test "converts empty directory" do
      assert {:ok, []} = Convert.directory_to_path([])
    end

    test "converts simple directory" do
      assert {:ok, ["path", "to", "dir"]} = Convert.directory_to_path(["path", "to", "dir"])
    end

    test "rejects directory with variable" do
      dir = ["path", KeyVal.variable([:int])]
      assert {:error, _} = Convert.directory_to_path(dir)
    end

    test "rejects directory with invalid element" do
      dir = ["path", 123]
      assert {:error, _} = Convert.directory_to_path(dir)
    end
  end

  describe "path_to_directory/1" do
    test "converts path to directory" do
      path = ["path", "to", "dir"]
      assert path == Convert.path_to_directory(path)
    end

    test "converts empty path" do
      assert [] == Convert.path_to_directory([])
    end
  end

  describe "has_versionstamp_future?/1" do
    test "returns false for tuple without versionstamp" do
      tuple = [KeyVal.int(1), KeyVal.int(2)]
      refute Convert.has_versionstamp_future?(tuple)
    end

    test "returns true for tuple with versionstamp_future" do
      tuple = [KeyVal.int(1), KeyVal.vstamp_future(0)]
      assert Convert.has_versionstamp_future?(tuple)
    end

    test "returns false for tuple with completed versionstamp" do
      tuple = [KeyVal.int(1), KeyVal.vstamp(<<0::80>>, 0)]
      refute Convert.has_versionstamp_future?(tuple)
    end

    test "finds versionstamp_future in nested tuple" do
      tuple = [[KeyVal.vstamp_future(0)], KeyVal.int(1)]
      assert Convert.has_versionstamp_future?(tuple)
    end
  end

  describe "versionstamp_position/1" do
    test "returns nil for tuple without versionstamp" do
      tuple = [KeyVal.int(1), KeyVal.int(2)]
      assert Convert.versionstamp_position(tuple) == nil
    end

    test "returns position of versionstamp_future" do
      tuple = [KeyVal.int(1), KeyVal.vstamp_future(0), KeyVal.int(3)]
      assert Convert.versionstamp_position(tuple) == 1
    end

    test "returns nil for completed versionstamp" do
      tuple = [KeyVal.int(1), KeyVal.vstamp(<<0::80>>, 0)]
      assert Convert.versionstamp_position(tuple) == nil
    end

    test "returns first versionstamp_future position" do
      tuple = [
        KeyVal.vstamp_future(0),
        KeyVal.int(1),
        KeyVal.vstamp_future(1)
      ]
      assert Convert.versionstamp_position(tuple) == 0
    end
  end
end
