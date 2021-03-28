package keyval

import (
	"encoding/binary"
	"math"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

func PackValue(val Value) ([]byte, error) {
	switch val.(type) {
	// Int
	case int64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(val.(int64)))
		return b, nil
	case int:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(val.(int)))
		return b, nil

	// Uint
	case uint64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, val.(uint64))
		return b, nil
	case uint:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(val.(uint)))
		return b, nil

	// Bool
	case bool:
		if val.(bool) {
			return []byte{1}, nil
		}
		return []byte{0}, nil

	// Float
	case float64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(val.(float64)))
		return b, nil
	case float32:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(float64(val.(float32))))
		return b, nil

	// String
	case string:
		return []byte(val.(string)), nil

	// Nil
	case nil:
		return nil, nil

	// Bytes
	case []byte:
		return val.([]byte), nil

	// UUID
	case UUID:
		uuid := val.(UUID)
		return uuid[:], nil

	// Tuple
	case Tuple:
		return ToFDBTuple(val.(Tuple)).Pack(), nil
	case tuple.Tuple:
		return val.(tuple.Tuple).Pack(), nil

	default:
		return nil, errors.Errorf("unknown Value type '%T'", val)
	}
}

func UnpackValue(typ ValueType, val []byte) (Value, error) {
	switch typ {
	case AnyType:
		return val, nil

	case IntType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return int64(binary.LittleEndian.Uint64(val)), nil

	case UintType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return binary.LittleEndian.Uint64(val), nil

	case BoolType:
		if len(val) != 1 {
			return nil, errors.New("not 1 byte")
		}
		if val[0] == 1 {
			return true, nil
		}
		return false, nil

	case FloatType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return math.Float64frombits(binary.LittleEndian.Uint64(val)), nil

	case StringType:
		return string(val), nil

	case BytesType:
		return val, nil

	case UUIDType:
		var uuid UUID
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
