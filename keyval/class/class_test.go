package class

import (
	"testing"

	q "github.com/janderland/fdbq/keyval"
	"github.com/stretchr/testify/assert"
)

func TestClassify(t *testing.T) {
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

func TestClassifyNil(t *testing.T) {
	tests := []struct {
		name string
		kv   q.KeyValue
	}{
		{
			name: "directory",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("hi"), nil, q.String("you")},
					Tuple:     nil,
				},
				Value: q.Bytes{},
			},
		},
		{
			name: "tuple",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("hi"), q.String("you")},
					Tuple:     q.Tuple{q.Int(34), q.String("wow"), nil},
				},
				Value: q.Bytes{},
			},
		},
		{
			name: "value",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("hi"), q.String("you")},
					Tuple:     q.Tuple{q.Int(34), q.String("wow")},
				},
				Value: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kind := Classify(test.kv)
			assert.Equal(t, Nil, kind)
		})
	}
}
