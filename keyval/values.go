package keyval

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

type Unpack func(val []byte) Value

func NewUnpack(query Value, order binary.ByteOrder) (Unpack, error) {
	if variable, isVar := query.(Variable); isVar {
		if len(variable) == 0 {
			return func(val []byte) Value {
				return val
			}, nil
		}

		return func(val []byte) Value {
			for _, typ := range variable {
				out, err := UnpackValue(order, typ, val)
				if err != nil {
					continue
				}
				return out
			}
			return nil
		}, nil
	} else {
		packed, err := PackValue(order, query)
		if err != nil {
			return nil, err
		}

		return func(val []byte) Value {
			if bytes.Equal(packed, val) {
				return query
			}
			return nil
		}, nil
	}
}

func PackValue(order binary.ByteOrder, val Value) ([]byte, error) {
	switch val := val.(type) {
	// Nil
	case nil:
		return nil, nil

	// Bool
	case bool:
		if val {
			return []byte{1}, nil
		}
		return []byte{0}, nil

	// Int
	case int64:
		b := make([]byte, 8)
		order.PutUint64(b, uint64(val))
		return b, nil
	case int:
		b := make([]byte, 8)
		order.PutUint64(b, uint64(val))
		return b, nil

	// Uint
	case uint64:
		b := make([]byte, 8)
		order.PutUint64(b, val)
		return b, nil
	case uint:
		b := make([]byte, 8)
		order.PutUint64(b, uint64(val))
		return b, nil

	// Float
	case float64:
		b := make([]byte, 8)
		order.PutUint64(b, math.Float64bits(val))
		return b, nil
	case float32:
		b := make([]byte, 8)
		order.PutUint64(b, math.Float64bits(float64(val)))
		return b, nil

	// String
	case string:
		return []byte(val), nil

	// Bytes
	case []byte:
		return val, nil

	// UUID
	case tuple.UUID:
		uuid := val
		return uuid[:], nil

	// Tuple
	case Tuple:
		return ToFDBTuple(val).Pack(), nil
	case tuple.Tuple:
		return val.Pack(), nil

	default:
		return nil, errors.Errorf("unknown Value type '%T'", val)
	}
}

func UnpackValue(order binary.ByteOrder, typ ValueType, val []byte) (Value, error) {
	switch typ {
	case AnyType:
		return val, nil

	case BoolType:
		if len(val) != 1 {
			return nil, errors.New("not 1 byte")
		}
		if val[0] == 1 {
			return true, nil
		}
		return false, nil

	case IntType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return int64(order.Uint64(val)), nil

	case UintType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return order.Uint64(val), nil

	case FloatType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return math.Float64frombits(order.Uint64(val)), nil

	case StringType:
		return string(val), nil

	case BytesType:
		return val, nil

	case UUIDType:
		var uuid tuple.UUID
		if n := copy(uuid[:], val); n != 16 {
			return nil, errors.New("not 16 bytes")
		}
		return uuid, nil

	case TupleType:
		tup, err := tuple.Unpack(val)
		return FromFDBTuple(tup), errors.Wrap(err, "failed to unpack tuple")

	default:
		return nil, errors.Errorf("unknown ValueType '%v'", typ)
	}
}
