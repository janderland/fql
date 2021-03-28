package keyval

import (
	"bytes"
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

// CompareTuples checks if the candidate Tuple conforms to the structure
// and values of the given pattern Tuple. The pattern Tuple may be of either
// ConstantKind, SingleReadKind, or RangeReadKind, while the candidate must
// be of ConstantKind. The elements of each Tuple are compared for equality.
// If an element of the pattern Tuple is a Variable, then the candidate's
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
func CompareTuples(pattern Tuple, candidate Tuple) []int {
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
	case MaybeMore:
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
	err := ReadTuple(candidate, AllowLong, func(iter *TupleIterator) error {
		for i, e := range pattern {
			switch e.(type) {
			// Nil
			case nil:
				if e != iter.Any() {
					index = []int{i}
					return nil
				}

			// Bool
			case bool:
				if iter.Bool() != e.(bool) {
					index = []int{i}
					return nil
				}

			// Int
			case int64:
				if iter.Int() != e.(int64) {
					index = []int{i}
					return nil
				}
			case int:
				if iter.Int() != int64(e.(int)) {
					index = []int{i}
					return nil
				}

			// Uint
			case uint64:
				if iter.Uint() != e.(uint64) {
					index = []int{i}
					return nil
				}
			case uint:
				if iter.Uint() != uint64(e.(uint)) {
					index = []int{i}
					return nil
				}

			// Float
			case float64:
				if iter.Float() != e.(float64) {
					index = []int{i}
					return nil
				}
			case float32:
				if iter.Float() != float64(e.(float32)) {
					index = []int{i}
					return nil
				}

			// big.Int
			case big.Int:
				v := e.(big.Int)
				if iter.BigInt().Cmp(&v) != 0 {
					index = []int{i}
					return nil
				}
			case *big.Int:
				if iter.BigInt().Cmp(e.(*big.Int)) != 0 {
					index = []int{i}
					return nil
				}

			// String
			case string:
				if iter.String() != e.(string) {
					index = []int{i}
					return nil
				}

			// Bytes
			case []byte:
				if bytes.Compare(iter.Bytes(), e.([]byte)) != 0 {
					index = []int{i}
					return nil
				}

			// UUID
			case UUID:
				if iter.UUID() != e.(UUID) {
					index = []int{i}
					return nil
				}

			// Tuple
			case Tuple:
				subIndex := CompareTuples(e.(Tuple), iter.Tuple())
				if len(subIndex) > 0 {
					index = append([]int{i}, subIndex...)
					return nil
				}
			case tuple.Tuple:
				subIndex := CompareTuples(FromFDBTuple(e.(tuple.Tuple)), iter.Tuple())
				if len(subIndex) > 0 {
					index = append([]int{i}, subIndex...)
					return nil
				}

			// Variable
			case Variable:
				// TODO: Check variable constraints.
				_ = iter.Any()
				break

			// Unknown
			default:
				index = []int{i}
				return nil
			}
		}
		return nil
	})
	if err != nil {
		if c, ok := err.(ConversionError); ok {
			return []int{c.Index}
		}
		panic(errors.Wrap(err, "unexpected error"))
	}
	return index
}
