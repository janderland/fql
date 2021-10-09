package class

import (
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/assert"
)

func TestKeyValue_Kind(t *testing.T) {
	tests := []struct {
		kind Class
		kv   q.KeyValue
	}{
		{
			kind: Constant,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Int(-38),
			},
		},
		{
			kind: Clear,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Clear{},
			},
		},
		{
			kind: SingleRead,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Variable{},
			},
		},
		{
			kind: RangeRead,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.Variable{}, q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Int(-38),
			},
		},
		{
			kind: RangeRead,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Variable{}, q.String("wow")},
				},
				Value: q.Nil{},
			},
		},
		{
			kind: RangeRead,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.Variable{}, q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Variable{},
			},
		},
		{
			kind: VariableClear,
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
			kind := Classify(test.kv)
			assert.Equal(t, test.kind, kind)
		})
	}
}
