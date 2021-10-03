package old

import (
	"testing"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/stretchr/testify/assert"
)

func TestToStringArray(t *testing.T) {
	dir := Directory{"my", "dir", "path"}
	arr, err := ToStringArray(dir)
	assert.NoError(t, err)
	assert.Equal(t, []string{"my", "dir", "path"}, arr)

	dir[1] = Variable{}
	_, err = ToStringArray(dir)
	assert.Error(t, err)
}

func TestFromStringArray(t *testing.T) {
	dir := FromStringArray([]string{"my", "dir", "path"})
	assert.Equal(t, Directory{"my", "dir", "path"}, dir)
}

func TestToFDBTuple(t *testing.T) {
	tup := ToFDBTuple(Tuple{nil, int64(22), false})
	assert.Equal(t, tuple.Tuple{nil, int64(22), false}, tup)

	tup = ToFDBTuple(Tuple{true, Tuple{32.8, "hi"}})
	assert.Equal(t, tuple.Tuple{true, tuple.Tuple{32.8, "hi"}}, tup)
}

func TestFromFDBTuple(t *testing.T) {
	tup := FromFDBTuple(tuple.Tuple{nil, int64(22), false})
	assert.Equal(t, Tuple{nil, int64(22), false}, tup)

	tup = FromFDBTuple(tuple.Tuple{true, Tuple{32.8, "hi"}})
	assert.Equal(t, Tuple{true, Tuple{32.8, "hi"}}, tup)
}

func TestSplitAtFirstVariable(t *testing.T) {
	prefix, variable, suffix := SplitAtFirstVariable([]interface{}{
		"one", int64(55), Variable{FloatType}, Tuple{-39.9},
	})
	assert.Equal(t, []interface{}{"one", int64(55)}, prefix)
	assert.Equal(t, &Variable{FloatType}, variable)
	assert.Equal(t, []interface{}{Tuple{-39.9}}, suffix)
}
