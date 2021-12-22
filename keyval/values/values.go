package values

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/pkg/errors"
)

type Filter func(val []byte) q.Value

func NewFilter(query q.Value, order binary.ByteOrder) (Filter, error) {
	if variable, ok := query.(q.Variable); ok {
		if len(variable) == 0 {
			return func(val []byte) q.Value {
				return q.Bytes(val)
			}, nil
		}

		return func(val []byte) q.Value {
			for _, typ := range variable {
				out, err := Unpack(val, typ, order)
				if err != nil {
					continue
				}
				return out
			}
			return nil
		}, nil
	} else {
		packed, err := Pack(query, order)
		if err != nil {
			return nil, err
		}

		return func(val []byte) q.Value {
			if bytes.Equal(packed, val) {
				return query
			}
			return nil
		}, nil
	}
}

func Pack(val q.Value, order binary.ByteOrder) ([]byte, error) {
	return newSerialization(order).Do(val)
}

func Unpack(val []byte, typ q.ValueType, order binary.ByteOrder) (q.Value, error) {
	switch typ {
	case q.AnyType:
		return q.Bytes(val), nil

	case q.BoolType:
		if len(val) != 1 {
			return nil, errors.New("not 1 byte")
		}
		if val[0] == 1 {
			return q.Bool(true), nil
		}
		return q.Bool(false), nil

	case q.IntType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return q.Int(order.Uint64(val)), nil

	case q.UintType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return q.Uint(order.Uint64(val)), nil

	case q.FloatType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return q.Float(math.Float64frombits(order.Uint64(val))), nil

	case q.StringType:
		return q.String(val), nil

	case q.BytesType:
		return q.Bytes(val), nil

	case q.UUIDType:
		var uuid q.UUID
		if n := copy(uuid[:], val); n != 16 {
			return nil, errors.New("not 16 bytes")
		}
		return uuid, nil

	case q.TupleType:
		tup, err := tuple.Unpack(val)
		return convert.FromFDBTuple(tup), errors.Wrap(err, "failed to unpack tuple")

	default:
		return nil, errors.Errorf("unknown ValueType '%v'", typ)
	}
}
