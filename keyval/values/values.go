package values

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/convert"
	"github.com/pkg/errors"
)

type Deserialize func(val []byte) (q.Value, error)

func NewDeserialize(query q.Value, order binary.ByteOrder, filter bool) (Deserialize, error) {
	if variable, ok := query.(q.Variable); ok {
		if len(variable) == 0 {
			return func(val []byte) (q.Value, error) {
				return q.Bytes(val), nil
			}, nil
		}

		return func(val []byte) (q.Value, error) {
			var errs []error
			for _, typ := range variable {
				out, err := Unpack(val, typ, order)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				return out, nil
			}
			if filter {
				return nil, nil
			}

			var str strings.Builder
			for i, err := range errs {
				if i > 0 {
					str.WriteRune(',')
				}
				str.WriteString(fmt.Sprintf("%s: %v", variable[i], err))
			}
			return nil, errors.Wrap(errors.New(str.String()), "failed to unpack as")
		}, nil
	} else {
		packed, err := Pack(query, order)
		if err != nil {
			return nil, errors.Wrap(err, "failed to pack query")
		}

		return func(val []byte) (q.Value, error) {
			if bytes.Equal(packed, val) {
				return query, nil
			}
			if filter {
				return nil, nil
			}
			return nil, errors.New("unexpected value")
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
