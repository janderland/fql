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

	nilFutureNil struct {
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
	_ fdb.FutureNil       = &nilFutureNil{}
)

// NewNilRange returns a nil implementation of fdb.Range
// where every operation is a no-op.
func NewNilRange() fdb.Range {
	return &nilRange{}
}

// NewNilSelectable returns a nil implementation of fdb.Selectable
// where every operation is a no-op.
func NewNilSelectable() fdb.Selectable {
	return &nilSelectable{}
}

// NewNilExactRange returns a nil implementation of fdb.ExactRange
// where every operation is a no-op.
func NewNilExactRange() fdb.ExactRange {
	return &nilExactRange{NewNilRange()}
}

// NewNilKeyConvertible returns a nil implementation of fdb.KeyConvertible
// where every operation is a no-op.
func NewNilKeyConvertible() fdb.KeyConvertible {
	return &nilKeyConvertible{}
}

// NewNilFuture returns a nil implementation of fdb.Future
// where every operation is a no-op.
func NewNilFuture() fdb.Future {
	return &nilFuture{}
}

// NewNilFutureByteSlice returns a nil implementation of fdb.FutureByteSlice
// where every operation is a no-op.
func NewNilFutureByteSlice() fdb.FutureByteSlice {
	return &nilFutureByteSlice{NewNilFuture()}
}

// NewNilFutureNil returns a nil implementation of fdb.FutureNil
// where every operation is a no-op.
func NewNilFutureNil() fdb.FutureNil {
	return &nilFutureNil{NewNilFuture()}
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

func (x *nilFutureNil) Get() error {
	return nil
}

func (x *nilFutureNil) MustGet() {
}
