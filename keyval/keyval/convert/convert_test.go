package convert

import (
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval/keyval"
	"github.com/stretchr/testify/assert"
)

func TestToStringArray(t *testing.T) {
	dir := q.Directory{q.String("my"), q.String("dir"), q.String("path")}
	arr, err := ToStringArray(dir)
	assert.NoError(t, err)
	assert.Equal(t, []string{"my", "dir", "path"}, arr)

	dir[1] = q.Variable{}
	_, err = ToStringArray(dir)
	assert.Error(t, err)
}

func TestFromStringArray(t *testing.T) {
	dir := FromStringArray([]string{"my", "dir", "path"})
	assert.Equal(t, q.Directory{q.String("my"), q.String("dir"), q.String("path")}, dir)
}

func TestToFDBTuple(t *testing.T) {
	tup, err := ToFDBTuple(q.Tuple{q.Nil{}, q.Int(22), q.Bool(false)})
	assert.NoError(t, err)
	assert.Equal(t, tuple.Tuple{nil, int64(22), false}, tup)

	tup, err = ToFDBTuple(q.Tuple{q.Bool(true), q.Tuple{q.Float(32.8), q.String("hi")}})
	assert.NoError(t, err)
	assert.Equal(t, tuple.Tuple{true, tuple.Tuple{32.8, "hi"}}, tup)
}

func TestFromFDBTuple(t *testing.T) {
	tup := FromFDBTuple(tuple.Tuple{nil, int64(22), false})
	assert.Equal(t, q.Tuple{q.Nil{}, q.Int(22), q.Bool(false)}, tup)

	tup = FromFDBTuple(tuple.Tuple{true, tuple.Tuple{32.8, "hi"}})
	assert.Equal(t, q.Tuple{q.Bool(true), q.Tuple{q.Float(32.8), q.String("hi")}}, tup)
}

func TestSplitAtFirstVariable(t *testing.T) {
	prefix, variable, suffix := SplitAtFirstVariable([]interface{}{
		"one", int64(55), q.Variable{q.FloatType}, q.Tuple{q.Float(-39.9)},
	})
	assert.Equal(t, []interface{}{"one", int64(55)}, prefix)
	assert.Equal(t, &q.Variable{q.FloatType}, variable)
	assert.Equal(t, []interface{}{q.Tuple{q.Float(-39.9)}}, suffix)
}
