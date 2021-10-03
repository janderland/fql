package convert

import (
	"math/big"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/apple/foundationdb/bindings/go/src/fdb/tuple"
	q "github.com/janderland/fdbq/keyval"
	"github.com/pkg/errors"
)

// ToStringArray attempts to convert a Directory to a string
// array. If the Directory contains non-string elements, an
// error is returned.
func ToStringArray(in q.Directory) ([]string, error) {
	out := make([]string, len(in))

	for i, element := range in {
		switch e := element.(type) {
		case q.String:
			out[i] = string(e)
		default:
			return nil, errors.Errorf("index '%d' has type '%T'", i, e)
		}
	}

	return out, nil
}

// FromStringArray converts a string array into a Directory.
func FromStringArray(in []string) q.Directory {
	out := make(q.Directory, len(in))

	for i := range in {
		out[i] = q.String(in[i])
	}

	return out
}

// ToFDBTuple converts a Tuple into a tuple.Tuple. Note that
// the resultant tuple.Tuple will be invalid if the original
// Tuple contains a Variable.
func ToFDBTuple(in q.Tuple) (tuple.Tuple, error) {
	out := make(tuple.Tuple, len(in))

	for i, element := range in {
		conv := conversion{}
		element.TupElement(&conv)
		if conv.err != nil {
			return nil, errors.Wrapf(conv.err, "failed to convert index %d", i)
		}
		out[i] = conv.out
	}

	return out, nil
}

// FromFDBTuple converts a tuple.Tuple into a Tuple.
func FromFDBTuple(in tuple.Tuple) q.Tuple {
	out := make(q.Tuple, len(in))

	for i, element := range in {
		switch element := element.(type) {
		case tuple.Tuple:
			out[i] = FromFDBTuple(element)
		default:
			out[i] = FromFDBElement(element)
		}
	}

	return out
}

func FromFDBElement(in tuple.TupleElement) q.TupElement {
	switch in := in.(type) {
	case []byte:
		return q.Bytes(in)
	case fdb.KeyConvertible:
		return q.Bytes(in.FDBKey())

	case string:
		return q.String(in)

	case int64:
		return q.Int(in)
	case int:
		return q.Int(in)

	case uint64:
		return q.Uint(in)
	case uint:
		return q.Uint(in)

	case big.Int:
		return q.BigInt(in)
	case *big.Int:
		return q.BigInt(*in)

	case float64:
		return q.Float(in)
	case float32:
		return q.Float(in)

	case bool:
		return q.Bool(in)

	case tuple.UUID:
		return q.UUID(in)

	case tuple.Tuple:
		return FromFDBTuple(in)

	case nil:
		return q.Nil{}

	default:
		panic(errors.Errorf("cannot convert type %T", in))
	}
}

// SplitAtFirstVariable accepts either a Directory or Tuple and returns a slice of the elements
// before the first variable, the first variable, and a slice of the elements after the variable.
// TODO: How should this method work with both tuples & directories.
func SplitAtFirstVariable(list []interface{}) ([]interface{}, *q.Variable, []interface{}) {
	for i, element := range list {
		if variable, ok := element.(q.Variable); ok {
			return list[:i], &variable, list[i+1:]
		}
	}
	return list, nil, nil
}

// RemoveMaybeMore removes a MaybeMore if it exists as the last element of the given Tuple.
func RemoveMaybeMore(tup q.Tuple) q.Tuple {
	if len(tup) > 0 {
		last := len(tup) - 1
		if _, ok := tup[last].(q.MaybeMore); ok {
			tup = tup[:last]
		}
	}
	return tup
}
