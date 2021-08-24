package compare

import (
	"bytes"
	"math/big"

	q "github.com/janderland/fdbq/keyval/keyval"
	"github.com/pkg/errors"
)

// Tuples checks if the candidate Tuple conforms to the structure
// and values of the given pattern Tuple. The pattern Tuple may contain
// Variable or MaybeMore while the candidate must not contain Variable or
// MaybeMore. The elements of each Tuple are compared for equality. If an
// element of the pattern Tuple is a Variable, then the candidate's
// corresponding element must conform for the constraints of the Variable.
// If all the elements match then a nil array is returned. If an element
// doesn't match, then an array is returned specifying the index path to the
// first mismatching element. For instance, given the following candidate
// tuple...
//
//   Tuple{55, Tuple{"hello", "world", Tuple{67}}}
//
// ...if the element with value "67" didn't match then the returned array
// would be []int{1,2,0}. If the Tuples aren't the same length, then the
// length of the shorter Tuple is returned as the sole element of the array.
func Tuples(pattern q.Tuple, candidate q.Tuple) []int {
	// Guards against invalid indexes in the
	// MaybeMore type switch.
	if len(pattern) == 0 {
		return []int{0}
	}

	// If this pattern ends with a MaybeMore, we
	// don't need to check if the candidate is
	// longer. Get rid of the MaybeMore as it's
	// not used in the actual comparisons.
	switch pattern[len(pattern)-1].(type) {
	case q.MaybeMore:
		pattern = pattern[:len(pattern)-1]
	default:
		if len(pattern) < len(candidate) {
			return []int{len(pattern)}
		}
	}

	// This check would be done by the ReadTuple()
	// call below, but as an optimization we may
	// exit early by checking here.
	if len(pattern) > len(candidate) {
		return []int{len(candidate)}
	}

	// Loop over both tuples, comparing their elements. If a pair of elements
	// don't match, place the current index in the array. If the comparison
	// happened within a sub-tuple, the index of the sub-tuple will be prepended
	// before the int of the mismatch within the tuple.
	var index []int
	err := q.ReadTuple(candidate, q.AllowLong, func(iter *q.TupleIterator) error {
		for i, e := range pattern {
			v := newVisitor(iter, i)
			if index = v.Visit(e); index != nil {
				return nil
			}
		}
		return nil
	})
	if err != nil {
		if c, ok := err.(q.ConversionError); ok {
			return []int{c.Index}
		}
		panic(errors.Wrap(err, "unexpected error"))
	}
	return index
}

type visitor struct {
	iter  *q.TupleIterator
	i     int
	index []int
}

func newVisitor(iter *q.TupleIterator, i int) visitor {
	return visitor{iter: iter, i: i}
}

func (x *visitor) Visit(e q.TupElement) []int {
	e.TupElement(x)
	return x.index
}

func (x *visitor) VisitNil(e q.Nil) {
	if e != x.iter.Any() {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitBool(e q.Bool) {
	if x.iter.MustBool() != e {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitInt(e q.Int) {
	if x.iter.MustInt() != e {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitUint(e q.Uint) {
	if x.iter.MustUint() != e {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitFloat(e q.Float) {
	if x.iter.MustFloat() != e {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitBigInt(e q.BigInt) {
	i1 := big.Int(x.iter.MustBigInt())
	i2 := big.Int(e)
	if i1.Cmp(&i2) != 0 {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitString(e q.String) {
	if x.iter.MustString() != e {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitBytes(e q.Bytes) {
	if !bytes.Equal(x.iter.MustBytes(), e) {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitUUID(e q.UUID) {
	if x.iter.MustUUID() != e {
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitTuple(e q.Tuple) {
	subIndex := Tuples(e, x.iter.MustTuple())
	if len(subIndex) > 0 {
		x.index = append([]int{x.i}, subIndex...)
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
		x.index = []int{x.i}
	}
}

func (x *visitor) VisitMaybeMore(_ q.MaybeMore) {
	x.index = []int{x.i}
}
