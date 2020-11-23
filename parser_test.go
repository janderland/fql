package fdbq

import (
	"testing"

	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/stretchr/testify/assert"
)

func TestParseKey(t *testing.T) {
	key, err := ParseKey("")
	assert.Error(t, err)
	assert.Nil(t, key)

	key, err = ParseKey("/my/dir")
	assert.NoError(t, err)
	assert.Equal(t, &Key{
		Directory: []string{"my", "dir"},
	}, key)

	key, err = ParseKey("('str', -13, (12e6))")
	assert.NoError(t, err)
	assert.Equal(t, &Key{
		Tuple: tup.Tuple{"str", int64(-13), tup.Tuple{12e6}},
	}, key)

	key, err = ParseKey("/my/dir('str', -13, (12e6))")
	assert.NoError(t, err)
	assert.Equal(t, &Key{
		Directory: []string{"my", "dir"},
		Tuple:     tup.Tuple{"str", int64(-13), tup.Tuple{12e6}},
	}, key)
}

func TestParseDirectory(t *testing.T) {
	dir, err := ParseDirectory("")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("/")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("hello")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("/hello")
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello"}, dir)

	dir, err = ParseDirectory("/hello/")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("/hello/world")
	assert.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, dir)

	dir, err = ParseDirectory("/hello/world/")
	assert.Error(t, err)
	assert.Nil(t, dir)
}

func TestParseTuple(t *testing.T) {
	tuple, err := ParseTuple("")
	assert.Error(t, err)
	assert.Nil(t, tuple)

	tuple, err = ParseTuple("(")
	assert.Error(t, err)
	assert.Nil(t, tuple)

	tuple, err = ParseTuple(")")
	assert.Error(t, err)
	assert.Nil(t, tuple)

	tuple, err = ParseTuple("()")
	assert.NoError(t, err)
	assert.Equal(t, tup.Tuple{}, tuple)

	tuple, err = ParseTuple("(17)")
	assert.NoError(t, err)
	assert.Equal(t, tup.Tuple{int64(17)}, tuple)

	tuple, err = ParseTuple("(17, 'hello world')")
	assert.NoError(t, err)
	assert.Equal(t, tup.Tuple{int64(17), "hello world"}, tuple)

	tuple, err = ParseTuple("('hello', 23.3, (-3))")
	assert.NoError(t, err)
	assert.Equal(t, tup.Tuple{"hello", 23.3, tup.Tuple{int64(-3)}}, tuple)

	tuple, err = ParseTuple("((bcefd2ec-4df5-43b6-8c79-81b70b886af9))")
	assert.NoError(t, err)
	assert.Equal(t, tup.Tuple{tup.Tuple{tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}, tuple)

	tuple, err = ParseTuple("('hello',, -3)")
	assert.Error(t, err)
	assert.Nil(t, tuple)
}

func TestParseData(t *testing.T) {
	data, err := ParseData("")
	assert.Error(t, err)
	assert.Nil(t, data)

	data, err = ParseData("nil")
	assert.NoError(t, err)
	assert.Nil(t, data)

	data, err = ParseData("true")
	assert.NoError(t, err)
	assert.Equal(t, true, data)

	data, err = ParseData("false")
	assert.NoError(t, err)
	assert.Equal(t, false, data)

	data, err = ParseData("'hello world'")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", data)

	data, err = ParseData("bcefd2ec-4df5-43b6-8c79-81b70b886af9")
	assert.NoError(t, err)
	assert.Equal(t, tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}, data)

	data, err = ParseData("123")
	assert.NoError(t, err)
	assert.Equal(t, int64(123), data)

	data, err = ParseData("-94.2")
	assert.NoError(t, err)
	assert.Equal(t, -94.2, data)

	data, err = ParseData("3.47e-8")
	assert.NoError(t, err)
	assert.Equal(t, 3.47e-8, data)

	data, err = ParseData("invalid")
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestParseString(t *testing.T) {
	str, err := ParseString("")
	assert.Error(t, err)
	assert.Equal(t, "", str)

	str, err = ParseString("'hello")
	assert.Error(t, err)
	assert.Equal(t, "", str)

	str, err = ParseString("'hello world'")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", str)
}

func TestParseUUID(t *testing.T) {
	id, err := ParseUUID("")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcec-4d-43b-8c-81b886af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefdyec-4df5-43%6-8c79-81b70bg86af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefd2ec-4df5-43b6-8c79-81b70b886af9")
	assert.NoError(t, err)
	assert.Equal(t, tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}, id)
}

func TestParseNumber(t *testing.T) {
	num, err := ParseNumber("")
	assert.Error(t, err)
	assert.Nil(t, num)

	num, err = ParseNumber("-34000")
	assert.NoError(t, err)
	assert.Equal(t, int64(-34000), num)

	num, err = ParseNumber("18446744073709551610")
	assert.NoError(t, err)
	assert.Equal(t, uint64(18446744073709551610), num)

	num, err = ParseNumber("94.33")
	assert.NoError(t, err)
	assert.Equal(t, 94.33, num)

	num, err = ParseNumber("12.54e-8")
	assert.NoError(t, err)
	assert.Equal(t, 12.54e-8, num)

	num, err = ParseNumber("hello")
	assert.Error(t, err)
	assert.Nil(t, num)
}
