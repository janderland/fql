package keyval

import (
	"math/big"
	"strconv"

	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	"github.com/pkg/errors"
)

// Kind categorizes a KeyValue.
type Kind string

const (
	// InvalidKind specifies that the KeyValue is malformed.
	InvalidKind Kind = ""

	// ConstantKind specifies that the KeyValue has no Variable,
	// MaybeMore, or Clear. This kind of KeyValue can be used to
	// perform a set operation or is returned by a get operation.
	ConstantKind Kind = "constant"

	// ClearKind specifies that the KeyValue has no Variable and
	// has a Clear Value. This kind of KeyValue can be used to
	// perform a clear operation.
	ClearKind Kind = "clear"

	// SingleReadKind specifies that the KeyValue has a Variable
	// Value (doesn't have a Clear) and doesn't have a Variable
	// in it's Key. This kind of KeyValue can be used to perform
	// a get operation that returns a single KeyValue.
	SingleReadKind Kind = "single"

	// RangeReadKind specifies that the KeyValue has a Variable
	// in it's Key and doesn't have a Clear Value. This kind of
	// KeyValue can be used to perform a get operation that
	// returns multiple KeyValue.
	RangeReadKind Kind = "range"
)

// Kind returns the Kind of the given KeyValue. If the KeyValue
// is malformed then InvalidKind and a non-nil error are returned.
// For details on what a malformed KeyValue is, see the KeyValue,
// Key, Directory, Tuple, and Value documentation.
func (kv *KeyValue) Kind() (Kind, error) {
	keyKind, err := keySubKind(kv.Key)
	if err != nil {
		return InvalidKind, errors.Wrap(err, "key is invalid")
	}
	valKind, err := valSubKind(kv.Value)
	if err != nil {
		return InvalidKind, errors.Wrap(err, "value is invalid")
	}

	if keyKind == constantSubKind {
		if valKind == clearSubKind {
			return ClearKind, nil
		}
		if valKind == variableSubKind {
			return SingleReadKind, nil
		}
		return ConstantKind, nil
	} else {
		if valKind == clearSubKind {
			return InvalidKind, errors.New("variable key with clear value")
		}
		return RangeReadKind, nil
	}
}

// subKind categorizes the Key, Directory,
// Tuple, and Value within a KeyValue.
type subKind = int

const (
	// invalidSubKind specifies that the component is malformed.
	invalidSubKind subKind = iota

	// constantSubKind specifies that the component contains no
	// Variable, MaybeMore, or Clear.
	constantSubKind

	// variableSubKind specifies that the component contains a
	// Variable or MaybeMore.
	variableSubKind

	// clearSubKind specifies that the component contains a Clear.
	clearSubKind
)

func keySubKind(key Key) (subKind, error) {
	dirKind, err := dirSubKind(key.Directory)
	if err != nil {
		return invalidSubKind, errors.Wrap(err, "directory is invalid")
	}
	tupKind, err := tupSubKind(key.Tuple)
	if err != nil {
		return invalidSubKind, errors.Wrap(err, "tuple is invalid")
	}
	if dirKind == variableSubKind || tupKind == variableSubKind {
		return variableSubKind, nil
	}
	return constantSubKind, nil
}

func dirSubKind(dir Directory) (subKind, error) {
	kind := constantSubKind

	for i, e := range dir {
		switch e.(type) {
		case string:
			continue

		case Variable:
			kind = variableSubKind

		default:
			return invalidSubKind, errors.Errorf("%s element is type %T", ordinal(i), e)
		}
	}

	return kind, nil
}

func tupSubKind(tup Tuple) (subKind, error) {
	kind := constantSubKind
	var err error

	for i, e := range tup {
		switch e.(type) {
		// Nil
		case nil:
			continue

		// Bool
		case bool:
			continue

		// Int
		case int64:
			continue
		case int:
			continue

		// Uint
		case uint64:
			continue
		case uint:
			continue

		// Float
		case float64:
			continue
		case float32:
			continue

		// BigInt
		case big.Int:
			continue
		case *big.Int:
			continue

		// String
		case string:
			continue

		case []byte:
			continue

		// UUID
		case UUID:
			continue

		// Tuple
		case Tuple:
			kind, err = tupSubKind(e.(Tuple))
			if err != nil {
				return kind, errors.Wrapf(err, "invalid tuple at element %d", i)
			}
		case tuple.Tuple:
			kind, err = tupSubKind(FromFDBTuple(e.(tuple.Tuple)))
			if err != nil {
				return kind, errors.Wrapf(err, "invalid tuple at element %d", i)
			}

		// Variable
		case Variable:
			kind = variableSubKind
		case MaybeMore:
			kind = variableSubKind

		default:
			return invalidSubKind, errors.Errorf("%s element is type %T", ordinal(i), e)
		}
	}

	return kind, nil
}

func valSubKind(val Value) (subKind, error) {
	switch val.(type) {
	// Nil
	case nil:
		return constantSubKind, nil

	// Bool
	case bool:
		return constantSubKind, nil

	// Int
	case int64:
		return constantSubKind, nil
	case int:
		return constantSubKind, nil

	// Uint
	case uint64:
		return constantSubKind, nil
	case uint:
		return constantSubKind, nil

	// Float
	case float64:
		return constantSubKind, nil
	case float32:
		return constantSubKind, nil

	// String
	case string:
		return constantSubKind, nil

	// Bytes
	case []byte:
		return constantSubKind, nil

	// UUID
	case UUID:
		return constantSubKind, nil

	// Tuple
	case Tuple:
		kind, err := tupSubKind(val.(Tuple))
		return kind, errors.Wrap(err, "invalid tuple")
	case tuple.Tuple:
		kind, err := tupSubKind(FromFDBTuple(val.(tuple.Tuple)))
		return kind, errors.Wrap(err, "invalid tuple")

	// Variable
	case Variable:
		return variableSubKind, nil

	// Clear
	case Clear:
		return clearSubKind, nil

	default:
		return invalidSubKind, errors.Errorf("value has type %T", val)
	}
}

// TODO: Reuse from parser package.
func ordinal(x int) string {
	suffix := "th"
	switch x % 10 {
	case 1:
		if x%100 != 11 {
			suffix = "st"
		}
	case 2:
		if x%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if x%100 != 13 {
			suffix = "rd"
		}
	}
	return strconv.Itoa(x) + suffix
}
