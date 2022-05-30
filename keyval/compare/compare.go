package compare

import (
	q "github.com/janderland/fdbq/keyval"
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
// ...if the element with value `67` didn't match then the returned array
// would be `[]int{1,2,0}`. If the Tuples aren't the same length, then the
// length of the shorter Tuple is used as the mismatching index.
func Tuples(pattern q.Tuple, candidate q.Tuple) []int {
	// If the pattern is empty, the candidate must
	// be empty as well.
	if len(pattern) == 0 {
		if len(candidate) == 0 {
			return nil
		}
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

	// The candidate must be at least as long
	// as the pattern.
	if len(pattern) > len(candidate) {
		return []int{len(candidate)}
	}

	// Loop over both tuples, comparing their elements.
	for i, element := range pattern {
		c := comparison{i: i, candidate: candidate[i]}
		element.TupElement(&c)
		if c.out != nil {
			return c.out
		}
	}
	return nil
}
