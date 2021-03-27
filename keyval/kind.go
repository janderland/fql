package keyval

import "github.com/pkg/errors"

// Kind categorizes a KeyValue.
type Kind string

const (
	// ConstantKind specifies that the KeyValue has no Variable
	// or Clear. This kind of KeyValue can be used to perform a
	// set operation or is returned by a get operation.
	ConstantKind Kind = "constant"

	// ClearKind specifies that the KeyValue has no Variable and
	// has a Clear Value. This kind of KeyValue can be used to
	// perform a clear operation.
	ClearKind Kind = "clear"

	// SingleReadKind specifies that the KeyValue has a Variable
	// Value (doesn't have a Clear) and doesn't have a Variable
	// in it's Key. This kind of KeyValue can be used to perform
	// a get operation that returns a single KeyValue.
	SingleReadKind Kind = "read"

	// RangeReadKind specifies that the KeyValue has a Variable
	// in it's Key and doesn't have a Clear Value. This kind of
	// KeyValue can be used to perform a get operation that
	// returns multiple KeyValue.
	RangeReadKind Kind = "range"

	// InvalidKind specifies the KeyValue is invalid.
	InvalidKind Kind = ""
)

func (kv *KeyValue) Kind() (Kind, error) {
	// TODO: Check validity of all elements.
	if KeyHasVariable(kv.Key) {
		if ValHasClear(kv.Value) {
			return InvalidKind, errors.New("cannot have variable key with clear value")
		}
		return RangeReadKind, nil
	} else {
		if ValHasClear(kv.Value) {
			return ClearKind, nil
		}
		if ValHasVariable(kv.Value) {
			return SingleReadKind, nil
		}
		return ConstantKind, nil
	}
}

// KeyHasVariable returns true if the Key contains a Variable or MaybeMore.
func KeyHasVariable(key Key) bool {
	if DirHasVariable(key.Directory) {
		return true
	}
	if TupHasVariable(key.Tuple) {
		return true
	}
	return false
}

// DirHasVariable returns true if the Directory contains a Variable.
func DirHasVariable(dir Directory) bool {
	for _, element := range dir {
		if _, ok := element.(Variable); ok {
			return true
		}
	}
	return false
}

// TupHasVariable returns true if the Tuple contains a Variable or MaybeMore.
func TupHasVariable(tup Tuple) bool {
	for _, element := range tup {
		switch element.(type) {
		case Tuple:
			if TupHasVariable(element.(Tuple)) {
				return true
			}
		case Variable:
			return true
		case MaybeMore:
			return true
		}
	}
	return false
}

// ValHasVariable returns true if the Value contains a Variable.
func ValHasVariable(val Value) bool {
	switch val.(type) {
	case Tuple:
		return TupHasVariable(val.(Tuple))
	case Variable:
		return true
	}
	return false
}

// ValHasClear returns true if the Value is an instance of Clear.
func ValHasClear(val Value) bool {
	_, hasClear := val.(Clear)
	return hasClear
}
