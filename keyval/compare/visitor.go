package compare

import (
	"bytes"
	"math/big"

	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

type comparison struct {
	candidate q.TupElement
	index     int

	// The index path to the first mismatching elements.
	firstMismatch []int
}

func (x *comparison) VisitNil(e q.Nil) {
	if _, ok := x.candidate.(q.Nil); !ok {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitBool(e q.Bool) {
	val, ok := x.candidate.(q.Bool)
	if !ok || val != e {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitInt(e q.Int) {
	val, ok := x.candidate.(q.Int)
	if !ok || val != e {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitUint(e q.Uint) {
	val, ok := x.candidate.(q.Uint)
	if !ok || val != e {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitFloat(e q.Float) {
	val, ok := x.candidate.(q.Float)
	if !ok || val != e {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitBigInt(e q.BigInt) {
	val, ok := x.candidate.(q.BigInt)
	if !ok {
		x.firstMismatch = []int{x.index}
	}

	biVal := big.Int(val)
	biE := big.Int(e)
	if biVal.Cmp(&biE) != 0 {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitString(e q.String) {
	val, ok := x.candidate.(q.String)
	if !ok || val != e {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitBytes(e q.Bytes) {
	val, ok := x.candidate.(q.Bytes)
	if !ok || bytes.Compare(val, e) != 0 {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitUUID(e q.UUID) {
	val, ok := x.candidate.(q.UUID)
	if !ok || val != e {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitTuple(e q.Tuple) {
	val, ok := x.candidate.(q.Tuple)
	if !ok {
		x.firstMismatch = []int{x.index}
	}

	subIndex := Tuples(e, val)
	if len(subIndex) > 0 {
		x.firstMismatch = append([]int{x.index}, subIndex...)
	}
}

func (x *comparison) VisitVariable(e q.Variable) {
	// An empty variable is equivalent
	// to an AnyType variable.
	if len(e) == 0 {
		return
	}

	found := false
loop:
	for _, vType := range e {
		switch vType {
		case q.AnyType:
			found = true
			break loop

		case q.IntType:
			if _, ok := x.candidate.(q.Int); ok {
				found = true
				break loop
			}

		case q.UintType:
			if _, ok := x.candidate.(q.Uint); ok {
				found = true
				break loop
			}

		case q.BoolType:
			if _, ok := x.candidate.(q.Bool); ok {
				found = true
				break loop
			}

		case q.FloatType:
			if _, ok := x.candidate.(q.Float); ok {
				found = true
				break loop
			}

		case q.BigIntType:
			if _, ok := x.candidate.(q.BigInt); ok {
				found = true
				break loop
			}

		case q.StringType:
			if _, ok := x.candidate.(q.String); ok {
				found = true
				break loop
			}

		case q.BytesType:
			if _, ok := x.candidate.(q.Bytes); ok {
				found = true
				break loop
			}

		case q.UUIDType:
			if _, ok := x.candidate.(q.UUID); ok {
				found = true
				break loop
			}

		case q.TupleType:
			if _, ok := x.candidate.(q.Tuple); ok {
				found = true
				break loop
			}

		default:
			panic(errors.Errorf("unrecognized variable type '%v'", vType))
		}
	}
	if !found {
		x.firstMismatch = []int{x.index}
	}
}

func (x *comparison) VisitMaybeMore(_ q.MaybeMore) {
	// By the time the visitor is used, the Tuples function
	// should have removed the trailing MaybeMore. So, any
	// MaybeMore we encounter here is invalid.
	x.firstMismatch = []int{x.index}
}
