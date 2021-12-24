package internal

import (
	"bytes"
	"encoding/binary"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/keyval/values"
	"github.com/pkg/errors"
)

type (
	// ValHandler processes the FDB values during a FDBQ read operation.
	ValHandler interface {
		Handle([]byte) (q.Value, error)
	}

	// pass is a ValHandler which passes the []byte through
	// as keyval.Bytes. It never returns an error.
	pass struct{}

	// unpack is a ValHandler which attempts to deserialize the given
	// []byte into one of the types specified in the variable. When
	// the first successful deserialization occurs, the resultant
	// value is returned. If the bytes cannot be deserialized to
	// any of the given types and filter=false, then an error is
	// returned. If filter=true, errors are not returned.
	unpack struct {
		variable q.Variable
		order    binary.ByteOrder
		filter   bool
	}

	// compare is a ValHandler which compares the given []byte to
	// the packed bytes of the query. If the bytes match, the original
	// query is returned. If the bytes don't match and filter=false,
	// then an error is returned. If filter=true, errors are not
	// returned.
	compare struct {
		query  q.Value
		packed []byte
		filter bool
	}
)

func NewValueHandler(query q.Value, order binary.ByteOrder, filter bool) (ValHandler, error) {
	if variable, ok := query.(q.Variable); ok {
		if len(variable) == 0 {
			return &pass{}, nil
		}
		return &unpack{
			variable: variable,
			order:    order,
			filter:   filter,
		}, nil
	} else {
		packed, err := values.Pack(query, order)
		if err != nil {
			return nil, errors.Wrap(err, "failed to pack query")
		}
		return &compare{
			query:  query,
			packed: packed,
			filter: filter,
		}, nil
	}
}

func (x *pass) Handle(val []byte) (q.Value, error) {
	return q.Bytes(val), nil
}

func (x *unpack) Handle(val []byte) (q.Value, error) {
	for _, typ := range x.variable {
		out, err := values.Unpack(val, typ, x.order)
		if err != nil {
			continue
		}
		return out, nil
	}
	if x.filter {
		return nil, nil
	}
	return nil, errors.New("unexpected value")
}

func (x *compare) Handle(val []byte) (q.Value, error) {
	if bytes.Equal(x.packed, val) {
		return x.query, nil
	}
	if x.filter {
		return nil, nil
	}
	return nil, errors.New("unexpected value")
}
