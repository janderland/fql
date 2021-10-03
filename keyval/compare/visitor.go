package compare

import (
	"bytes"
	"math/big"

	q "github.com/janderland/fdbq/keyval"
	iter "github.com/janderland/fdbq/keyval/iterator"
	"github.com/pkg/errors"
)

type visitor struct {
	// The iterator wrapping the candidate tuple.
	iter *iter.TupleIterator

	// The index of the tuple element being compared.
	index int

	// The index path of the first mismatching element.
	mismatchIndexPath []int
}

func (x *visitor) VisitNil(e q.Nil) {
	if e != x.iter.Any() {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitBool(e q.Bool) {
	if x.iter.MustBool() != e {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitInt(e q.Int) {
	if x.iter.MustInt() != e {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitUint(e q.Uint) {
	if x.iter.MustUint() != e {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitFloat(e q.Float) {
	if x.iter.MustFloat() != e {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitBigInt(e q.BigInt) {
	i1 := big.Int(x.iter.MustBigInt())
	i2 := big.Int(e)
	if i1.Cmp(&i2) != 0 {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitString(e q.String) {
	if x.iter.MustString() != e {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitBytes(e q.Bytes) {
	if !bytes.Equal(x.iter.MustBytes(), e) {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitUUID(e q.UUID) {
	if x.iter.MustUUID() != e {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitTuple(e q.Tuple) {
	subIndex := Tuples(e, x.iter.MustTuple())
	if len(subIndex) > 0 {
		x.mismatchIndexPath = append([]int{x.index}, subIndex...)
	}
}

func (x *visitor) VisitVariable(e q.Variable) {
	// An empty variable is equivalent
	// to an AnyType variable.
	if len(e) == 0 {
		_ = x.iter.Any()
		return
	}

	found := false
	for _, vType := range e {
		var err error

		switch vType {
		case q.AnyType:
			_ = x.iter.Any()
		case q.IntType:
			_, err = x.iter.Int()
		case q.UintType:
			_, err = x.iter.Uint()
		case q.BoolType:
			_, err = x.iter.Bool()
		case q.FloatType:
			_, err = x.iter.Float()
		case q.BigIntType:
			_, err = x.iter.BigInt()
		case q.StringType:
			_, err = x.iter.String()
		case q.BytesType:
			_, err = x.iter.Bytes()
		case q.UUIDType:
			_, err = x.iter.UUID()
		case q.TupleType:
			_, err = x.iter.Tuple()
		default:
			panic(errors.Errorf("unrecognized variable type '%v'", vType))
		}

		if err == nil {
			found = true
			break
		}
	}
	if !found {
		x.mismatchIndexPath = []int{x.index}
	}
}

func (x *visitor) VisitMaybeMore(_ q.MaybeMore) {
	// By the time the visitor is used, the Tuples function
	// should have removed the trailing MaybeMore. So, any
	// MaybeMore we encounter here is invalid.
	x.mismatchIndexPath = []int{x.index}
}
