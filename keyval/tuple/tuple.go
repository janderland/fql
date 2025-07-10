package tuple

import (
	q "github.com/janderland/fql/keyval"
)

// TODO: Can some of the stream functions for analyzing
// tuples be moved here?

// TODO: Should this be a generic type-searching
// method instead?
func HasVStampFuture(tup q.Tuple) bool {
	for _, element := range tup {
		if _, ok := element.(q.VStampFuture); ok {
			return true
		}
		if subTup, ok := element.(q.Tuple); ok {
			if out := HasVStampFuture(subTup); out {
				return true
			}
		}
	}
	return false
}
