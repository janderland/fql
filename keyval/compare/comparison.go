package compare

import (
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

var _ q.TupleVisitor = &comparison{}

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

func (x *comparison) VisitNil(e q.Nil) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitBool(e q.Bool) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitInt(e q.Int) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitUint(e q.Uint) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitFloat(e q.Float) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitBigInt(e q.BigInt) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitString(e q.String) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitBytes(e q.Bytes) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitUUID(e q.UUID) {
	if !e.Eq(x.candidate) {
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitTuple(e q.Tuple) {
	val, ok := x.candidate.(q.Tuple)
	if !ok {
		x.mismatch = []int{x.i}
	}

	mismatch := Tuples(e, val)
	if len(mismatch) > 0 {
		x.mismatch = append([]int{x.i}, mismatch...)
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
		x.mismatch = []int{x.i}
	}
}

func (x *comparison) VisitMaybeMore(_ q.MaybeMore) {
	// By the time the visitor is used, the Tuples function
	// should have removed the trailing MaybeMore. So, any
	// MaybeMore we encounter here is invalid.
	x.mismatch = []int{x.i}
}
