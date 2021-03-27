package keyval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirIsVariable(t *testing.T) {
	dir := Directory{"my", "dir", "path"}
	assert.False(t, DirHasVariable(dir))
	dir[1] = Variable{}
	assert.True(t, DirHasVariable(dir))
}

func TestTupIsVariable(t *testing.T) {
	tup := Tuple{true, 22.8, "yup"}
	assert.False(t, TupHasVariable(tup))
	tup[1] = Variable{}
	assert.True(t, TupHasVariable(tup))
	tup[1] = Tuple{"hello", Variable{}}
	assert.True(t, TupHasVariable(tup))
}

func TestValIsVariable(t *testing.T) {
	val := Tuple{true, 22.8, "yup"}
	assert.False(t, ValHasVariable(val))
	val[1] = Variable{}
	assert.True(t, ValHasVariable(val))
	assert.True(t, ValHasVariable(Variable{}))
}

func TestIsVariable(t *testing.T) {
	newKV := func() *KeyValue {
		return &KeyValue{
			Key: Key{
				Directory: Directory{"my", "dir", "path"},
				Tuple:     Tuple{true, 22.8, "yup"},
			},
			Value: Tuple{true, 22.8, "yup"},
		}
	}
	assert.False(t, HasVariable(newKV()))

	kv := newKV()
	kv.Key.Directory[1] = Variable{}
	assert.True(t, HasVariable(kv))

	kv = newKV()
	kv.Key.Tuple[1] = Variable{}
	assert.True(t, HasVariable(kv))

	kv = newKV()
	kv.Value = Variable{}
	assert.True(t, HasVariable(kv))

	kv = newKV()
	kv.Value.(Tuple)[1] = Variable{}
	assert.True(t, HasVariable(kv))
}
