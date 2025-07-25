package class

import (
	"testing"

	"github.com/stretchr/testify/require"

	q "github.com/janderland/fql/keyval"
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
			kind: VStampKey,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.VStampFuture{UserVersion: 202}, q.String("wow")},
				},
				Value: q.Int(-38),
			},
		},
		{
			kind: VStampVal,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.String("wow")},
				},
				Value: q.Tuple{q.VStampFuture{UserVersion: 202}},
			},
		},
		{
			kind: VStampVal,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.String("wow")},
				},
				Value: q.VStampFuture{UserVersion: 202},
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
			kind: ReadSingle,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Variable{},
			},
		},
		{
			kind: ReadRange,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.Variable{}, q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Int(-38),
			},
		},
		{
			kind: ReadRange,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("my"), q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Variable{}, q.String("wow")},
				},
				Value: q.Nil{},
			},
		},
		{
			kind: ReadRange,
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.Variable{}, q.String("dir")},
					Tuple:     q.Tuple{q.Int(123), q.Float(-55.8), q.String("wow")},
				},
				Value: q.Variable{},
			},
		},
	}

	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			kind := Classify(test.kv)
			require.Equal(t, test.kind, kind)
		})
	}
}

func TestClassifyInvalid(t *testing.T) {
	tests := []struct {
		name string
		kv   q.KeyValue
	}{
		{
			name: "nil directory",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("hi"), nil, q.String("you")},
					Tuple:     nil,
				},
				Value: q.Bytes{},
			},
		},
		{
			name: "nil tuple",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("hi"), q.String("you")},
					Tuple:     q.Tuple{q.Int(34), q.String("wow"), nil},
				},
				Value: q.Bytes{},
			},
		},
		{
			name: "nil value",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("hi"), q.String("you")},
					Tuple:     q.Tuple{q.Int(34), q.String("wow")},
				},
				Value: nil,
			},
		},
		{
			name: "vstamp key & value",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("where"), q.String("how?")},
					Tuple: q.Tuple{
						q.Int(42),
						q.VStampFuture{UserVersion: 384},
					},
				},
				Value: q.Tuple{q.VStampFuture{UserVersion: 384}},
			},
		},
		{
			name: "vstamp double",
			kv: q.KeyValue{
				Key: q.Key{
					Directory: q.Directory{q.String("where"), q.String("how?")},
					Tuple: q.Tuple{
						q.VStampFuture{UserVersion: 384},
						q.VStampFuture{UserVersion: 400},
					},
				},
				Value: q.Nil{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kind := Classify(test.kv)
			require.Regexp(t, "invalid", string(kind))
		})
	}
}
