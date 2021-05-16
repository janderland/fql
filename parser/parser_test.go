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
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseKeyValue(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, *ast)

			str, err := FormatKeyValue(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
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
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseKeyValue(test.str)
			assert.Error(t, err)
			assert.Nil(t, ast)
		})
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
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseValue(test.str)
			assert.Error(t, err)
			assert.Nil(t, ast)
		})
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
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseDirectory(test.str)
			assert.Error(t, err)
			assert.Nil(t, ast)
		})
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
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseTuple(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)

			str, err := FormatTuple(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
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
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseTuple(test.str)
			assert.Error(t, err)
			assert.Nil(t, ast)
		})
	}
}

func TestParseData(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  interface{}
	}{
		{name: "nil", str: "nil", ast: nil},
		{name: "true", str: "true", ast: true},
		{name: "false", str: "false", ast: false},
		{name: "variable", str: "{int}", ast: keyval.Variable{keyval.IntType}},
		{name: "string", str: "\"hello world\"", ast: "hello world"},
		{name: "uuid", str: "bcefd2ec-4df5-43b6-8c79-81b70b886af9", ast: tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}},
		{name: "int", str: "123", ast: int64(123)},
		{name: "float", str: "-94.2", ast: -94.2},
		{name: "scientific", str: "3.47e-08", ast: 3.47e-8},
	}

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseData(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)

			str, err := FormatData(test.ast)
			assert.NoError(t, err)
			assert.Equal(t, test.str, str)
		})
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "invalid", str: "invalid"},
	}

	for _, test := range parseFailures {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseData(test.str)
			assert.Error(t, err)
			assert.Nil(t, ast)
		})
	}
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
	roundTrips := []struct {
		name string
		str  string
		ast  string
	}{
		{name: "regular", str: "\"hello world\"", ast: "hello world"},
	}

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseString(test.str)
			assert.NoError(t, err)
			assert.Equal(t, test.ast, ast)
		})
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "no close", str: "\"hello"},
		{name: "no open", str: "hello\""},
		{name: "single quote", str: "'hello'"},
	}

	for _, test := range parseFailures {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseString(test.str)
			assert.Error(t, err)
			assert.Empty(t, ast)
		})
	}
}

func TestParseUUID(t *testing.T) {
	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "bad group 1", str: "cefd2ec-4df5-43b6-8c79-81b70b886af9"},
		{name: "bad group 2", str: "bcefd2ec-df5-43b6-8c79-81b70b886af9"},
		{name: "bad group 3", str: "bcefd2ec-4df5-3b6-8c79-81b70b886af9"},
		{name: "bad group 4", str: "bcefd2ec-4df5-43b6-c79-81b70b886af9"},
		{name: "bad group 5", str: "bcefd2ec-4df5-43b6-8c79-1b70b886af9"},
		{name: "long", str: "bcefdyec-4df5-43%6-8c79-81b70bg86af9"},
	}

	for _, test := range parseFailures {
		ast, err := ParseUUID(test.str)
		assert.Error(t, err)
		assert.Equal(t, tup.UUID{}, ast)
	}

	roundTrips := []struct {
		name string
		str  string
		ast  tup.UUID
	}{
		{name: "normal", str: "bcefd2ec-4df5-43b6-8c79-81b70b886af9",
			ast: tup.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}},
	}

	for _, test := range roundTrips {
		ast, err := ParseUUID(test.str)
		assert.NoError(t, err)
		assert.Equal(t, test.ast, ast)

		str := FormatUUID(test.ast)
		assert.Equal(t, test.str, str)
	}
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
