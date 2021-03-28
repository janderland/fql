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
