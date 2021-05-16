package parser

import (
	"testing"

	tup "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

	"github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/assert"
)

func TestParseKeyValue(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  keyval.KeyValue
	}{
		{name: "full", str: "/hi/there(54,nil)=(33.8)",
			ast: keyval.KeyValue{Key: keyval.Key{Directory: keyval.Directory{"hi", "there"}, Tuple: keyval.Tuple{int64(54), nil}}, Value: keyval.Tuple{33.8}}},
	}

	for _, test := range roundTrips {
		ast, err := ParseKeyValue(test.str)
		assert.NoError(t, err)
		assert.Equal(t, test.ast, *ast)

		str, err := FormatKeyValue(test.ast)
		assert.NoError(t, err)
		assert.Equal(t, test.str, str)
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "empty tup", str: "()"},
		{name: "double sep", str: "()=()=()"},
		{name: "bad key", str: "badkey=()"},
		{name: "bad value", str: "()=badvalue"},
	}

	for _, test := range parseFailures {
		ast, err := ParseKeyValue(test.str)
		assert.Error(t, err)
		assert.Nil(t, ast)
	}
}

func TestParseKey(t *testing.T) {
	roundTrips := []struct {
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

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseKey(test.str)
			assert.NoError(t, err)
			assert.Equal(t, &test.ast, ast)

			str, err := FormatKey(*ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "bad dir", str: "baddir"},
		{name: "bad tup", str: "/dir(badtup"},
	}

	for _, fail := range parseFailures {
		t.Run(fail.name, func(t *testing.T) {
			key, err := ParseKey(fail.str)
			assert.Error(t, err)
			assert.Nil(t, key)
		})
	}
}

func TestParseValue(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  keyval.Value
	}{
		{name: "clear", str: "clear", ast: keyval.Clear{}},
		{name: "tuple", str: "(-16,13.2,\"hi\")", ast: keyval.Tuple{int64(-16), 13.2, "hi"}},
		{name: "raw", str: "-16", ast: int64(-16)},
	}

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseValue(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)

			str, err := FormatValue(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
	}

	for _, test := range parseFailures {
		ast, err := ParseValue(test.str)
		assert.Error(t, err)
		assert.Nil(t, ast)
	}
}

func TestParseDirectory(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  keyval.Directory
	}{
		{name: "single", str: "/hello", ast: keyval.Directory{"hello"}},
		{name: "multi", str: "/hello/world", ast: keyval.Directory{"hello", "world"}},
		{name: "variable", str: "/hello/{int}/thing", ast: keyval.Directory{"hello", keyval.Variable{keyval.IntType}, "thing"}},
	}

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseDirectory(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)

			str, err := FormatDirectory(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "no paths", str: "/"},
		{name: "no slash", str: "hello"},
		{name: "empty path", str: "/ /path"},
		{name: "trailing slash", str: "/hello/world/"},
		{name: "invalid var", str: "/hello/{/thing"},
	}

	for _, test := range parseFailures {
		ast, err := ParseDirectory(test.str)
		assert.Error(t, err)
		assert.Nil(t, ast)
	}
}

func TestParseTuple(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  keyval.Tuple
	}{
		{name: "empty", str: "()", ast: keyval.Tuple{}},
		{name: "one", str: "(17)", ast: keyval.Tuple{int64(17)}},
		{name: "two", str: "(17,\"hello world\")", ast: keyval.Tuple{int64(17), "hello world"}},
		{name: "sub tuple", str: "(\"hello\",23.3,(-3))", ast: keyval.Tuple{"hello", 23.3, keyval.Tuple{int64(-3)}}},
		{name: "uuid", str: "((bcefd2ec-4df5-43b6-8c79-81b70b886af9))",
			ast: keyval.Tuple{keyval.Tuple{tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},
	}

	for _, test := range roundTrips {
		ast, err := ParseTuple(test.str)
		assert.NoError(t, err)
		assert.Equal(t, test.ast, ast)

		str, err := FormatTuple(test.ast)
		assert.NoError(t, err)
		assert.Equal(t, test.str, str)
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "no close", str: "("},
		{name: "no open", str: ")"},
		{name: "bad element", str: "(bad)"},
		{name: "empty element", str: "(\"hello\",, -3)"},
	}

	for _, test := range parseFailures {
		ast, err := ParseTuple(test.str)
		assert.Error(t, err)
		assert.Nil(t, ast)
	}
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
