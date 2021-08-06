package keyval

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

type Unpacker struct {
	isVar    bool
	variable Variable
	unpacked Value
	packed   []byte
}

func NewUnpacker(query Value) (*Unpacker, error) {
	x := Unpacker{}
	var err error

	if x.variable, x.isVar = query.(Variable); !x.isVar {
		x.unpacked = query
		x.packed, err = PackValue(query)
		if err != nil {
			return nil, err
		}
	}
	return &x, nil
}

func (x *Unpacker) Unpack(val []byte) Value {
	if x.isVar {
		if len(x.variable) == 0 {
			return val
		}
		for _, typ := range x.variable {
			out, err := UnpackValue(typ, val)
			if err != nil {
				continue
			}
			return out
		}
		return nil
	} else {
		if bytes.Equal(x.packed, val) {
			return x.unpacked
		}
		return nil
	}
}

func PackValue(val Value) ([]byte, error) {
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
		binary.LittleEndian.PutUint64(b, uint64(val))
		return b, nil
	case int:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(val))
		return b, nil

	// Uint
	case uint64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, val)
		return b, nil
	case uint:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(val))
		return b, nil

	// Float
	case float64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(val))
		return b, nil
	case float32:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(float64(val)))
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

func UnpackValue(typ ValueType, val []byte) (Value, error) {
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
		return int64(binary.LittleEndian.Uint64(val)), nil

	case UintType:
		if len(val) != 8 {
			return nil, errors.New("not 8 bytes")
		}
		return binary.LittleEndian.Uint64(val), nil

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
