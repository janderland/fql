package parser

import (
	"testing"

	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

	"github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/assert"
)

func TestParseKeyValue(t *testing.T) {
	q, err := ParseKeyValue("")
	assert.Error(t, err)
	assert.Nil(t, q)

	q, err = ParseKeyValue("()")
	assert.Error(t, err)
	assert.Nil(t, q)

	q, err = ParseKeyValue("()=()=()")
	assert.Error(t, err)
	assert.Nil(t, q)

	q, err = ParseKeyValue("badkey=()")
	assert.Error(t, err)
	assert.Nil(t, q)

	q, err = ParseKeyValue("()=badvalue")
	assert.Error(t, err)
	assert.Nil(t, q)

	q, err = ParseKeyValue("()=()")
	assert.NoError(t, err)
	assert.Equal(t, &keyval.KeyValue{
		Key:   keyval.Key{Tuple: keyval.Tuple{}},
		Value: keyval.Tuple{},
	}, q)

	q, err = ParseKeyValue("() \t= \n()")
	assert.NoError(t, err)
	assert.Equal(t, &keyval.KeyValue{
		Key:   keyval.Key{Tuple: keyval.Tuple{}},
		Value: keyval.Tuple{},
	}, q)

	str, err := FormatKeyValue(keyval.KeyValue{
		Key: keyval.Key{
			Directory: keyval.Directory{"hi", "there"},
			Tuple:     keyval.Tuple{54, nil}},
		Value: keyval.Tuple{33.8},
	})
	assert.NoError(t, err)
	assert.Equal(t, "/hi/there(54,nil)=(33.8)", str)
}

func TestParseKey(t *testing.T) {
	tests := []struct {
		name string
		str  string
		ast  keyval.Key
	}{
		{name: "dir", str: "/my/dir",
			ast: keyval.Key{Directory: keyval.Directory{"my", "dir"}}},
		{name: "tup", str: "(\"str\",-13,(1.2e+13))",
			ast: keyval.Key{Tuple: keyval.Tuple{"str", int64(-13), keyval.Tuple{1.2e13}}}},
		{name: "full", str: "/my/dir(\"str\",-13,(1.2e+13))",
			ast: keyval.Key{Directory: keyval.Directory{"my", "dir"}, Tuple: keyval.Tuple{"str", int64(-13), keyval.Tuple{1.2e13}}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseKey(test.str)
			assert.NoError(t, err)
			assert.Equal(t, &test.ast, ast)

			str, err := FormatKey(*ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	fails := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "bad dir", str: "baddir"},
		{name: "bad tup", str: "/dir(badtup"},
	}

	for _, fail := range fails {
		t.Run(fail.name, func(t *testing.T) {
			key, err := ParseKey(fail.str)
			assert.Error(t, err)
			assert.Nil(t, key)
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name string
		str  string
		ast  keyval.Value
	}{
		{name: "clear", str: "clear", ast: keyval.Clear{}},
		{name: "tuple", str: "(-16,13.2,\"hi\")", ast: keyval.Tuple{int64(-16), 13.2, "hi"}},
		{name: "raw", str: "-16", ast: int64(-16)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseValue(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)

			str, err := FormatValue(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	val, err := ParseValue("")
	assert.Error(t, err)
	assert.Nil(t, val)
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

	dir, err = ParseDirectory("/ /empty-path")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("/hello/")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("/hello/world/")
	assert.Error(t, err)
	assert.Nil(t, dir)

	dir, err = ParseDirectory("/hello/{/thing")
	assert.Error(t, err)
	assert.Nil(t, dir)

	tests := []struct {
		name string
		str  string
		ast  keyval.Directory
	}{
		{name: "single", str: "/hello", ast: keyval.Directory{"hello"}},
		{name: "multi", str: "/hello/world", ast: keyval.Directory{"hello", "world"}},
		{name: "variable", str: "/hello/{int}/thing", ast: keyval.Directory{"hello", keyval.Variable{keyval.IntType}, "thing"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseDirectory(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)

			str, err := FormatDirectory(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	dir, err = ParseDirectory("/hello\n/ world")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Directory{"hello", "world"}, dir)
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
	assert.Equal(t, keyval.Tuple{}, tuple)

	tuple, err = ParseTuple("(badelem)")
	assert.Error(t, err)
	assert.Nil(t, tuple)

	tuple, err = ParseTuple("(17)")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Tuple{int64(17)}, tuple)

	tuple, err = ParseTuple("(17, \"hello world\")")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Tuple{int64(17), "hello world"}, tuple)

	tuple, err = ParseTuple("(\"hello\", 23.3, (-3))")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Tuple{"hello", 23.3, keyval.Tuple{int64(-3)}}, tuple)

	tuple, err = ParseTuple("((bcefd2ec-4df5-43b6-8c79-81b70b886af9))")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Tuple{keyval.Tuple{tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}, tuple)

	tuple, err = ParseTuple("(\"hello\",, -3)")
	assert.Error(t, err)
	assert.Nil(t, tuple)

	tuple, err = ParseTuple("(\n-15 \t, \n \"hello\"  )")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Tuple{int64(-15), "hello"}, tuple)
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

	data, err = ParseData("{int}")
	assert.NoError(t, err)
	assert.Equal(t, keyval.Variable{keyval.IntType}, data)

	data, err = ParseData("\"hello world\"")
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

func TestParseVariable(t *testing.T) {
	tests := []struct {
		name string
		str  string
		ast  keyval.Variable
	}{
		{name: "empty", str: "{}", ast: nil},
		{name: "single", str: "{int}", ast: keyval.Variable{keyval.IntType}},
		{name: "multiple", str: "{int|float|tuple}", ast: keyval.Variable{keyval.IntType, keyval.FloatType, keyval.TupleType}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseVariable(test.str)
			assert.Equal(t, test.ast, ast)
			assert.NoError(t, err)

			str := FormatVariable(test.ast)
			assert.Equal(t, test.str, str)
		})
	}

	fails := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "unclosed", str: "{"},
		{name: "unopened", str: "}"},
		{name: "invalid", str: "{invalid}"},
	}

	for _, test := range fails {
		t.Run(test.name, func(t *testing.T) {
			v, err := ParseVariable(test.str)
			assert.Error(t, err)
			assert.Nil(t, v)
		})
	}
}

func TestParseString(t *testing.T) {
	str, err := ParseString("")
	assert.Error(t, err)
	assert.Equal(t, "", str)

	str, err = ParseString("\"hello")
	assert.Error(t, err)
	assert.Equal(t, "", str)

	str, err = ParseString("'hello")
	assert.Error(t, err)
	assert.Equal(t, "", str)

	str, err = ParseString("\"hello world\"")
	assert.NoError(t, err)
	assert.Equal(t, "hello world", str)

	str = FormatString("hello world")
	assert.Equal(t, "\"hello world\"", str)
}

func TestParseUUID(t *testing.T) {
	id, err := ParseUUID("")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("cefd2ec-4df5-43b6-8c79-81b70b886af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefd2ec-df5-43b6-8c79-81b70b886af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefd2ec-4df5-3b6-8c79-81b70b886af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefd2ec-4df5-43b6-c79-81b70b886af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefd2ec-4df5-43b6-8c79-1b70b886af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefdyec-4df5-43%6-8c79-81b70bg86af9")
	assert.Error(t, err)
	assert.Equal(t, tup.UUID{}, id)

	id, err = ParseUUID("bcefd2ec-4df5-43b6-8c79-81b70b886af9")
	assert.NoError(t, err)
	assert.Equal(t, tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}, id)

	str := FormatUUID(tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9})
	assert.Equal(t, "bcefd2ec-4df5-43b6-8c79-81b70b886af9", str)
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
