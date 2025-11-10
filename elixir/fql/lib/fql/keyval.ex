defmodule Fql.KeyVal do
  @moduledoc """
  Core data structures representing key-values and related utilities.

  These types model both queries and the data returned by queries. They can be
  constructed from query strings using `Fql.Parser` but are also designed to be
  easily constructed directly in Elixir code.

  ## Embedded Query Strings

  Instead of using string literals for queries (like SQL), FQL allows programmers
  to directly construct queries using the types in this module, allowing some
  syntax errors to be caught at compile time.

  ## Types

  The main types in this module are:

  - `query/0` - Union type for all query types
  - `key_value/0` - A key-value pair
  - `key/0` - A key (directory + tuple)
  - `directory/0` - A directory path
  - `tuple/0` - A tuple of elements
  - `value/0` - A value (primitive, tuple, variable, or clear)
  - `variable/0` - A placeholder for schema matching
  - `maybe_more/0` - Allows prefix matching in tuples

  ## Primitive Types

  The primitive types include: nil, int, uint, bool, float, string, uuid, bytes,
  vstamp, and vstamp_future.
  """

  # Query types
  @type query :: key_value() | key() | directory()

  @type key_value :: %{
    key: key(),
    value: value()
  }

  @type key :: %{
    directory: directory(),
    tuple: tuple()
  }

  @type directory :: [dir_element()]
  @type dir_element :: String.t() | variable()

  @type tuple :: [tup_element()]

  @type tup_element ::
    tuple()
    | nil_type()
    | integer()  # Int or Uint
    | boolean()
    | float()
    | String.t()
    | uuid()
    | binary()
    | variable()
    | maybe_more()
    | vstamp()
    | vstamp_future()

  @type value ::
    tuple()
    | nil_type()
    | integer()  # Int or Uint (tagged)
    | boolean()
    | float()
    | String.t()
    | uuid()
    | binary()
    | variable()
    | clear()
    | vstamp()
    | vstamp_future()

  @type variable :: %{
    types: [value_type()]
  }

  @type maybe_more :: :maybe_more

  @type clear :: :clear

  # Primitive types (using tagged tuples for disambiguation)
  @type nil_type :: :nil
  @type int :: {:int, integer()}
  @type uint :: {:uint, non_neg_integer()}
  @type uuid :: {:uuid, <<_::128>>}
  @type bytes :: {:bytes, binary()}

  @type vstamp :: %{
    tx_version: <<_::80>>,  # 10 bytes
    user_version: non_neg_integer()  # uint16
  }

  @type vstamp_future :: %{
    user_version: non_neg_integer()  # uint16
  }

  # Value types for variables
  @type value_type ::
    :any
    | :int
    | :uint
    | :bool
    | :float
    | :string
    | :bytes
    | :uuid
    | :tuple
    | :vstamp

  @doc """
  Returns all valid value types.
  """
  @spec all_types() :: [value_type()]
  def all_types do
    [:any, :int, :uint, :bool, :float, :string, :bytes, :uuid, :tuple, :vstamp]
  end

  @doc """
  Creates a KeyValue struct.
  """
  @spec key_value(key(), value()) :: key_value()
  def key_value(key, value) do
    %{key: key, value: value}
  end

  @doc """
  Creates a Key struct.
  """
  @spec key(directory(), tuple()) :: key()
  def key(directory, tuple) do
    %{directory: directory, tuple: tuple}
  end

  @doc """
  Creates a Variable with the given types.
  """
  @spec variable([value_type()]) :: variable()
  def variable(types \\ []) do
    %{types: types}
  end

  @doc """
  Creates a VStamp.
  """
  @spec vstamp(<<_::80>>, non_neg_integer()) :: vstamp()
  def vstamp(tx_version, user_version) do
    %{tx_version: tx_version, user_version: user_version}
  end

  @doc """
  Creates a VStampFuture.
  """
  @spec vstamp_future(non_neg_integer()) :: vstamp_future()
  def vstamp_future(user_version) do
    %{user_version: user_version}
  end

  @doc """
  Creates an Int value (tagged tuple).
  """
  @spec int(integer()) :: int()
  def int(value), do: {:int, value}

  @doc """
  Creates a Uint value (tagged tuple).
  """
  @spec uint(non_neg_integer()) :: uint()
  def uint(value), do: {:uint, value}

  @doc """
  Creates a UUID value (tagged tuple).
  """
  @spec uuid(<<_::128>>) :: uuid()
  def uuid(value), do: {:uuid, value}

  @doc """
  Creates a Bytes value (tagged tuple).
  """
  @spec bytes(binary()) :: bytes()
  def bytes(value), do: {:bytes, value}
end
