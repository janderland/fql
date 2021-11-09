package compare

import (
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

var _ q.TupleOperation = &comparison{}

type comparison struct {
	i         int
	candidate q.TupElement
	mismatch  []int
}

func newComparison(i int, candidate q.TupElement) comparison {
	return comparison{i: i, candidate: candidate}
}

func (x comparison) Do(pattern q.TupElement) []int {
	pattern.TupElement(&x)
	return x.mismatch
}

func (x *comparison) ForNil(e q.Nil) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForBool(e q.Bool) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForInt(e q.Int) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForUint(e q.Uint) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForFloat(e q.Float) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForBigInt(e q.BigInt) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForString(e q.String) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForBytes(e q.Bytes) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForUUID(e q.UUID) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForTuple(e q.Tuple) {
	val, ok := x.candidate.(q.Tuple)
	if !ok {
		x.mismatch = []int{x.i}
	}

	mismatch := Tuples(e, val)
	if len(mismatch) > 0 {
		x.mismatch = append([]int{x.i}, mismatch...)
	}
}

func (x *comparison) ForVariable(e q.Variable) {
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
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) ForMaybeMore(_ q.MaybeMore) {
	// By the time the visitor is used, the Tuples function
	// should have removed the trailing MaybeMore. So, any
	// MaybeMore we encounter here is invalid.
	x.mismatch = []int{x.i}
}
