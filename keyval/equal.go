package keyval

import (
	"bytes"
	"math/big"
)

func (x Nil) Eq(e interface{}) bool {
	_, ok := e.(Nil)
	return ok
}

func (x Int) Eq(e interface{}) bool {
	return x == e
}

func (x Uint) Eq(e interface{}) bool {
	return x == e
}

func (x Bool) Eq(e interface{}) bool {
	return x == e
}

func (x Float) Eq(e interface{}) bool {
	return x == e
}

func (x BigInt) Eq(e interface{}) bool {
	v, ok := e.(BigInt)
	if !ok {
		return false
	}

	X, E := big.Int(x), big.Int(v)
	return X.Cmp(&E) == 0
}

func (x String) Eq(e interface{}) bool {
	return x == e
}

func (x UUID) Eq(e interface{}) bool {
	return x == e
}

func (x Bytes) Eq(e interface{}) bool {
	v, ok := e.(Bytes)
	return ok && bytes.Equal(x, v)
}

func (x Variable) Eq(e interface{}) bool {
	v, ok := e.(Variable)
	if !ok {
		return false
	}
	if len(x) != len(v) {
		return false
	}
	for i := range x {
		if x[i] != v[i] {
			return false
		}
	}
	return true
}

func (x MaybeMore) Eq(e interface{}) bool {
	_, ok := e.(MaybeMore)
	return ok
}

func (x Tuple) Eq(e interface{}) bool {
	v, ok := e.(Tuple)
	if !ok {
		return false
	}
	if len(x) != len(v) {
		return false
	}
	for i := range x {
		if !x[i].Eq(v[i]) {
			return false
		}
	}
	return true
}

func (x Clear) Eq(e interface{}) bool {
	_, ok := e.(Clear)
	return ok
}
