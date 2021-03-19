package keyval

import (
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

func CompareTuples(pattern Tuple, candidate Tuple) ([]int, error) {
	// Guards against invalid indexes in the
	// MaybeMore type switch.
	if len(pattern) == 0 {
		return nil, errors.New("empty pattern")
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
			return []int{len(pattern) + 1}, nil
		}
	}

	// This check would be done by the ReadTuple()
	// call below, but as an optimization we may
	// exit early by checking here.
	if len(pattern) > len(candidate) {
		return []int{len(candidate) + 1}, nil
	}

	// Loop over both tuples, comparing their elements. If a pair of elements
	// don't match, place the current index in the array. If the comparison
	// happened within a sub-tuple, the index of the sub-tuple will be prepended
	// before the int of the mismatch within the tuple.
	var index []int
	err := ReadTuple(candidate, AllowLong, func(iter *TupleIterator) error {
		for i, e := range pattern {
			switch e.(type) {
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

			// String
			case string:
				if iter.String() != e.(string) {
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

			// Bool
			case bool:
				if iter.Bool() != e.(bool) {
					index = []int{i}
					return nil
				}

			// Nil
			case nil:
				if e != iter.Any() {
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

			// UUID
			case UUID:
				if iter.UUID() != e.(UUID) {
					index = []int{i}
					return nil
				}

			// Variable
			case Variable:
				// TODO: Check variable constraints.
				_ = iter.Any()
				break

			// Tuple
			case Tuple:
				subIndex, err := CompareTuples(e.(Tuple), iter.Tuple())
				if err != nil {
					return errors.Wrap(err, "failed to compare sub-tuple")
				}
				if len(subIndex) > 0 {
					index = append([]int{i}, subIndex...)
					return nil
				}
			case tuple.Tuple:
				subIndex, err := CompareTuples(FromFDBTuple(e.(tuple.Tuple)), iter.Tuple())
				if err != nil {
					return errors.Wrap(err, "failed to compare sub-tuple")
				}
				if len(subIndex) > 0 {
					index = append([]int{i}, subIndex...)
					return nil
				}

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
			return []int{c.Index}, nil
		}
		return nil, errors.Wrap(err, "unexpected error")
	}
	return index, nil
}

/*
func CompareVariables(pattern Value, candidate Value) ([]int, error) {
	switch pattern.(type) {
	case Variable:
		// TODO: Check variable constraints.
		return nil, nil

	// Tuple
	case Tuple, tuple.Tuple:
		switch candidate.(type) {
		case Tuple:

		}
		subIndex, err := CompareTuples(pattern.(Tuple))
		if err != nil {
			return errors.Wrap(err, "failed to compare sub-tuple")
		}
		if len(subIndex) > 0 {
			index = append([]int{i}, subIndex...)
			return nil
		}
	}
}
*/
