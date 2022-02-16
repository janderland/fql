package parser

import (
	"strings"
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/require"
)

func TestDirectory(t *testing.T) {
	RoundTripTests{
		{name: "single", str: "/hello", ast: q.Directory{q.String("hello")}},
		{name: "multi", str: "/hello/world", ast: q.Directory{q.String("hello"), q.String("world")}},
		{name: "variable", str: "/hello/<>/thing", ast: q.Directory{q.String("hello"), q.Variable{}, q.String("thing")}},
	}.run(t)

	ParseFailureTests{
		{name: "empty", str: ""},
		{name: "no paths", str: "/"},
		{name: "no slash", str: "hello"},
		{name: "trailing slash", str: "/hello/world/"},
		{name: "invalid var", str: "/hello/</thing"},
	}.run(t)
}

/*
func TestTuple(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  q.Tuple
	}{
		{name: "empty",
			str: "{}",
			ast: q.Tuple{}},

		{name: "one",
			str: "{17}",
			ast: q.Tuple{q.Int(17)}},

		{name: "two",
			str: "{17,\"hello world\"}",
			ast: q.Tuple{q.Int(17), q.String("hello world")}},

		{name: "sub tuple",
			str: "{\"hello\",23.3,{-3}}",
			ast: q.Tuple{q.String("hello"), q.Float(23.3), q.Tuple{q.Int(-3)}}},

		{name: "uuid",
			str: "{{bcefd2ec-4df5-43b6-8c79-81b70b886af9}}",
			ast: q.Tuple{q.Tuple{q.UUID{0xbc, 0xef, 0xd2, 0xec, 0x4d, 0xf5, 0x43, 0xb6, 0x8c, 0x79, 0x81, 0xb7, 0x0b, 0x88, 0x6a, 0xf9}}}},

		{name: "maybe more",
			str: "{18.2,0xffaa,...}",
			ast: q.Tuple{q.Float(18.2), q.Bytes{0xFF, 0xAA}, q.MaybeMore{}}},
	}

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseTuple(test.str)
			require.NoError(t, err)
			require.Equal(t, test.ast, ast)
		})
	}

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "no close", str: "{"},
		{name: "no open", str: "}"},
		{name: "bad element", str: "{bad}"},
		{name: "empty element", str: "{\"hello\",, -3}"},
	}

	for _, test := range parseFailures {
		t.Run(test.name, func(t *testing.T) {
			ast, err := ParseTuple(test.str)
			require.Error(t, err)
			require.Nil(t, ast)
		})
	}
}
*/

type RoundTripTest struct {
	name string
	str  string
	ast  q.Query
}

type RoundTripTests []RoundTripTest

func (x RoundTripTests) run(t *testing.T) {
	for _, test := range x {
		t.Run(test.name, func(t *testing.T) {
			p := NewParser(NewScanner(strings.NewReader(test.str)))
			ast, err := p.Parse()
			require.NoError(t, err)
			require.Equal(t, test.ast, ast)

			/*
				TODO: Enable test.
				str, err := FormatDirectory(test.ast)
				require.NoError(t, err)
				require.Equal(t, test.str, str)
			*/
		})
	}
}

type ParseFailureTest struct {
	name string
	str  string
}

type ParseFailureTests []ParseFailureTest

func (x ParseFailureTests) run(t *testing.T) {
	for _, test := range x {
		t.Run(test.name, func(t *testing.T) {
			p := NewParser(NewScanner(strings.NewReader(test.str)))
			ast, err := p.Parse()
			require.Error(t, err)
			require.Nil(t, ast)
		})
	}
}
