package parser

import (
	"strings"
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/require"
)

func TestDirectory(t *testing.T) {
	roundTrips := []struct {
		name string
		str  string
		ast  q.Query
	}{
		{name: "single", str: "/hello", ast: q.Directory{q.String("hello")}},
		{name: "multi", str: "/hello/world", ast: q.Directory{q.String("hello"), q.String("world")}},
		{name: "variable", str: "/hello/<>/thing", ast: q.Directory{q.String("hello"), q.Variable{}, q.String("thing")}},
	}

	for _, test := range roundTrips {
		t.Run(test.name, func(t *testing.T) {
			ast, err := Parse(NewScanner(strings.NewReader(test.str)))
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

	parseFailures := []struct {
		name string
		str  string
	}{
		{name: "empty", str: ""},
		{name: "no paths", str: "/"},
		{name: "no slash", str: "hello"},
		{name: "trailing slash", str: "/hello/world/"},
		{name: "invalid var", str: "/hello/</thing"},
	}

	for _, test := range parseFailures {
		t.Run(test.name, func(t *testing.T) {
			ast, err := Parse(NewScanner(strings.NewReader(test.str)))
			require.Error(t, err)
			require.Nil(t, ast)
		})
	}
}
