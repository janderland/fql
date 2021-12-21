package facade

import (
	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

type (
	nilRange struct{}

	nilSelectable struct{}

	nilExactRange struct {
		fdb.Range
	}

	nilKeyConvertible struct{}

	nilFuture struct{}

	nilFutureByteSlice struct {
		fdb.Future
	}
)

var (
	_ fdb.Range          = &nilRange{}
	_ fdb.Selectable     = &nilSelectable{}
	_ fdb.ExactRange     = &nilExactRange{}
	_ fdb.KeyConvertible = &nilKeyConvertible{}

	_ fdb.Future          = &nilFuture{}
	_ fdb.FutureByteSlice = &nilFutureByteSlice{}
)

func NewNilRange() fdb.Range {
	return &nilRange{}
}

func NewNilSelectable() fdb.Selectable {
	return &nilSelectable{}
}

func NewNilExactRange() fdb.ExactRange {
	return &nilExactRange{NewNilRange()}
}

func NewNilKeyConvertible() fdb.KeyConvertible {
	return &nilKeyConvertible{}
}

func NewNilFuture() fdb.Future {
	return &nilFuture{}
}

func NewNilFutureByteSlice() fdb.FutureByteSlice {
	return &nilFutureByteSlice{NewNilFuture()}
}

func (x *nilFuture) BlockUntilReady() {}

func (x *nilFuture) IsReady() bool {
	return false
}

func (x *nilFuture) Cancel() {}

func (x *nilSelectable) FDBKeySelector() fdb.KeySelector {
	return fdb.KeySelector{}
}

func (x *nilRange) FDBRangeKeySelectors() (begin, end fdb.Selectable) {
	return NewNilSelectable(), nil
}

func (x *nilExactRange) FDBRangeKeys() (begin, end fdb.KeyConvertible) {
	return NewNilKeyConvertible(), NewNilKeyConvertible()
}

func (x *nilKeyConvertible) FDBKey() fdb.Key {
	return nil
}

func (x *nilFutureByteSlice) Get() ([]byte, error) {
	return nil, nil
}

func (x *nilFutureByteSlice) MustGet() []byte {
	return nil
}
