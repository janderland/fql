defmodule Fql.KeyVal.Values do
  @moduledoc """
  Serializes and deserializes values for FoundationDB storage.
  """

  alias Fql.KeyVal

  @type byte_order :: :big | :little

  @doc """
  Serializes a KeyVal.Value into a byte string for writing to the DB.
  """
  @spec pack(KeyVal.value(), byte_order(), boolean()) :: {:ok, binary()} | {:error, String.t()}
  def pack(nil, _order, _vstamp), do: {:error, "value cannot be nil"}

  def pack(:nil, _order, _vstamp), do: {:ok, <<>>}

  def pack(val, order, _vstamp) when is_boolean(val) do
    {:ok, if(val, do: <<1>>, else: <<0>>)}
  end

  def pack({:int, val}, order, _vstamp) when is_integer(val) do
    bytes = encode_int64(val, order)
    {:ok, bytes}
  end

  def pack({:uint, val}, order, _vstamp) when is_integer(val) and val >= 0 do
    bytes = encode_uint64(val, order)
    {:ok, bytes}
  end

  def pack(val, order, _vstamp) when is_float(val) do
    bytes = encode_float64(val, order)
    {:ok, bytes}
  end

  def pack(val, _order, _vstamp) when is_binary(val) do
    {:ok, val}
  end

  def pack({:bytes, val}, _order, _vstamp) when is_binary(val) do
    {:ok, val}
  end

  def pack({:uuid, val}, _order, _vstamp) when byte_size(val) == 16 do
    {:ok, val}
  end

  def pack(%{tx_version: tx_ver, user_version: user_ver}, _order, _vstamp) do
    # VStamp
    user_bytes = <<user_ver::little-unsigned-16>>
    {:ok, tx_ver <> user_bytes}
  end

  def pack(val, _order, true) when is_list(val) do
    # Tuple - would need FDB tuple packing
    {:error, "tuple packing not yet implemented"}
  end

  def pack(val, _order, _vstamp) when is_list(val) do
    {:error, "tuple values require vstamp=true"}
  end

  def pack(_val, _order, _vstamp) do
    {:error, "unsupported value type"}
  end

  @doc """
  Deserializes a KeyVal.Value from a byte string read from the DB.
  """
  @spec unpack(binary(), KeyVal.value_type(), byte_order()) ::
    {:ok, KeyVal.value()} | {:error, String.t()}
  def unpack(val, :any, _order) do
    {:ok, {:bytes, val}}
  end

  def unpack(val, :bool, _order) do
    case byte_size(val) do
      1 ->
        case val do
          <<1>> -> {:ok, true}
          <<0>> -> {:ok, false}
          _ -> {:error, "invalid bool value"}
        end
      _ ->
        {:error, "bool must be 1 byte"}
    end
  end

  def unpack(val, :int, order) do
    case byte_size(val) do
      8 ->
        int_val = decode_int64(val, order)
        {:ok, {:int, int_val}}
      _ ->
        {:error, "int must be 8 bytes"}
    end
  end

  def unpack(val, :uint, order) do
    case byte_size(val) do
      8 ->
        uint_val = decode_uint64(val, order)
        {:ok, {:uint, uint_val}}
      _ ->
        {:error, "uint must be 8 bytes"}
    end
  end

  def unpack(val, :float, order) do
    case byte_size(val) do
      8 ->
        float_val = decode_float64(val, order)
        {:ok, float_val}
      _ ->
        {:error, "float must be 8 bytes"}
    end
  end

  def unpack(val, :string, _order) do
    {:ok, val}
  end

  def unpack(val, :bytes, _order) do
    {:ok, {:bytes, val}}
  end

  def unpack(val, :uuid, _order) do
    case byte_size(val) do
      16 -> {:ok, {:uuid, val}}
      _ -> {:error, "uuid must be 16 bytes"}
    end
  end

  def unpack(val, :tuple, _order) do
    # Would need FDB tuple unpacking
    {:error, "tuple unpacking not yet implemented"}
  end

  def unpack(val, :vstamp, _order) do
    case byte_size(val) do
      12 ->
        <<tx_ver::binary-size(10), user_ver::little-unsigned-16>> = val
        {:ok, %{tx_version: tx_ver, user_version: user_ver}}
      _ ->
        {:error, "vstamp must be 12 bytes"}
    end
  end

  def unpack(_val, type, _order) do
    {:error, "unknown value type: #{inspect(type)}"}
  end

  # Encoding helpers
  defp encode_int64(val, :big), do: <<val::big-signed-64>>
  defp encode_int64(val, :little), do: <<val::little-signed-64>>

  defp encode_uint64(val, :big), do: <<val::big-unsigned-64>>
  defp encode_uint64(val, :little), do: <<val::little-unsigned-64>>

  defp encode_float64(val, :big), do: <<val::big-float-64>>
  defp encode_float64(val, :little), do: <<val::little-float-64>>

  # Decoding helpers
  defp decode_int64(<<val::big-signed-64>>, :big), do: val
  defp decode_int64(<<val::little-signed-64>>, :little), do: val

  defp decode_uint64(<<val::big-unsigned-64>>, :big), do: val
  defp decode_uint64(<<val::little-unsigned-64>>, :little), do: val

  defp decode_float64(<<val::big-float-64>>, :big), do: val
  defp decode_float64(<<val::little-float-64>>, :little), do: val
end
