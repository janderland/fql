package facade

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/subspace"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
)

type nilSubspace struct {
	fdb.ExactRange
	fdb.KeyConvertible
}

var _ subspace.Subspace = &nilSubspace{}

func NewNilSubspace() subspace.Subspace {
	return &nilSubspace{NewNilExactRange(), NewNilKeyConvertible()}
}

func (x *nilSubspace) Sub(...tuple.TupleElement) subspace.Subspace {
	return x
}

func (x *nilSubspace) Bytes() []byte {
	return nil
}

func (x *nilSubspace) Pack(_ tuple.Tuple) fdb.Key {
	return nil
}

func (x *nilSubspace) Unpack(_ fdb.KeyConvertible) (tuple.Tuple, error) {
	return nil, nil
}

func (x *nilSubspace) Contains(_ fdb.KeyConvertible) bool {
	return false
}
