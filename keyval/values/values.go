// Package values provides functions for serializing and deserializing
// keyval.Value being written or read from the DB.
package values

import (
	"encoding/binary"
	"math"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/pkg/errors"
)

// Pack serializes keyval.Value into a bytes string for writing to the DB.
func Pack(val q.Value, order binary.ByteOrder) ([]byte, error) {
	if val == nil {
		return nil, errors.New("value cannot be nil")
	}
	s := serialization{order: order}
	val.Value(&s)
	return s.out, s.err
}

// Unpack deserializes keyval.Value from a byte string read from the DB.
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
