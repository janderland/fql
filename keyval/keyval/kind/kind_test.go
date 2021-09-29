package kind

import (
	"testing"

	q "github.com/janderland/fdbq/keyval/keyval"
	"github.com/stretchr/testify/assert"
)

func TestKeyValue_Kind(t *testing.T) {
	tests := []struct {
		kind Kind
		kv   q.KeyValue
	}{
		{
			kind: ConstantKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Int(-38),
			},
		},
		{
			kind: ClearKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Clear{},
			},
		},
		{
			kind: SingleReadKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Variable{},
			},
		},
		{
			kind: RangeReadKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.Variable{}, q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Int(-38),
			},
		},
		{
			kind: RangeReadKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Variable{}, q.String("wow")},
				},
				Value: q.Nil{},
			},
		},
		{
			kind: RangeReadKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.Variable{}, q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Variable{},
			},
		},
		{
			kind: VariableClearKind,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.MaybeMore{}},
				},
				Value: q.Clear{},
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			kind := Which(test.kv)
			assert.Equal(t, test.kind, kind)
		})
	}
}
