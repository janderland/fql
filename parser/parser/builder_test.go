package parser

import (
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/require"
)

func TestTupBuilder_startSubTuple(t *testing.T) {
	var b tupBuilder

	b.startSubTuple()
	require.Equal(t, q.Tuple{q.Tuple{}}, b.get())

	b.startSubTuple()
	require.Equal(t, q.Tuple{q.Tuple{q.Tuple{}}}, b.get())
}

func TestTupBuilder_appendStringToTuple(t *testing.T) {
	var b tupBuilder

	b.appendToTuple(q.String(""))
	require.Equal(t, q.Tuple{q.String("")}, b.get())

	b.startSubTuple()
	b.appendToTuple(q.String(""))
	require.Equal(t, q.Tuple{q.String(""), q.Tuple{q.String("")}}, b.get())
}

func TestTupBuilder_appendToLastTupElem(t *testing.T) {
	var b tupBuilder

	b.startSubTuple()
	b.startSubTuple()
	b.appendToTuple(q.String(""))
	b.appendToLastTupElem("hello")
	require.Equal(t, q.Tuple{q.Tuple{q.Tuple{q.String("hello")}}}, b.get())
}
