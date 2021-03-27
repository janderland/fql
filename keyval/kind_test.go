package keyval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyValue_Kind(t *testing.T) {
	tests := []struct {
		kind Kind
		kv   KeyValue
	}{
		{
			kind: ConstantKind,
			kv: KeyValue{
				Key: Key{
					Directory: Directory{"my", "dir"},
					Tuple:     Tuple{123, -55.8, "wow"},
				},
				Value: -38,
			},
		},
		{
			kind: ClearKind,
			kv: KeyValue{
				Key: Key{
					Directory: Directory{"my", "dir"},
					Tuple:     Tuple{123, -55.8, "wow"},
				},
				Value: Clear{},
			},
		},
		{
			kind: SingleReadKind,
			kv: KeyValue{
				Key: Key{
					Directory: Directory{"my", "dir"},
					Tuple:     Tuple{123, -55.8, "wow"},
				},
				Value: Variable{},
			},
		},
		{
			kind: RangeReadKind,
			kv: KeyValue{
				Key: Key{
					Directory: Directory{Variable{}, "dir"},
					Tuple:     Tuple{123, -55.8, "wow"},
				},
				Value: -38,
			},
		},
		{
			kind: RangeReadKind,
			kv: KeyValue{
				Key: Key{
					Directory: Directory{Variable{}, "dir"},
					Tuple:     Tuple{123, -55.8, "wow"},
				},
				Value: Variable{},
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			kind, err := test.kv.Kind()
			assert.NoError(t, err)
			assert.Equal(t, test.kind, kind)
		})
	}
}

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

func TestKeyHasVariable(t *testing.T) {
	newKey := func() Key {
		return Key{
			Directory: Directory{"my", "dir", "path"},
			Tuple:     Tuple{true, 22.8, "yup"},
		}
	}
	assert.False(t, KeyHasVariable(newKey()))

	key := newKey()
	key.Directory[1] = Variable{}
	assert.True(t, KeyHasVariable(key))

	key = newKey()
	key.Tuple[1] = Variable{}
	assert.True(t, KeyHasVariable(key))
}
