// Package keyval provides the key value data structure
// and functions for inspecting the key values.
package keyval

import "github.com/apple/foundationdb/bindings/go/src/fdb/tuple"

type (
	// A KeyValue is a query or result depending on the
	// context. If the KeyValue is a result, it will not
	// contain a Variable.
	KeyValue struct {
		Key   Key
		Value Value
	}

	// A Key represents an FDB key made up of a Directory
	// and optionally a Tuple. A Key cannot have both an
	// empty Directory and an empty Tuple.
	Key struct {
		Directory Directory
		Tuple     Tuple
	}

	// A Directory is equivalent to a path used by the
	// directory layer of the FDB API. A Directory may
	// contain instances of string or Variable.
	Directory []interface{}

	// A Tuple is similar to a tuple.Tuple. It may contain
	// anything in a valid tuple.Tuple in addition to
	// Variable and Tuple.
	Tuple []interface{}

	// A Value represents an FDB value stored alongside
	// a key. This type may contain nil, bool, string,
	// uint64, float64, UUID, Tuple, Variable, or Clear.
	Value interface{}

	// A UUID is equivalent to a tuple.UUID.
	UUID = tuple.UUID

	// A Variable is used as a placeholder for any valid
	// values within the contexts that allow it.
	Variable struct {
		Type []ValueType
	}

	// A MaybeMore is a special kind of Tuple element. It
	// may only appear as the last element of the Tuple.
	// It designates that the Tuple will match all Tuples
	// which contain a matching prefix.
	MaybeMore struct{}

	// A Clear is a special kind of Value which designates
	// a KeyValue as a clear operation.
	Clear struct{}
)

// ValueType specifies the variable's expected type.
type ValueType string

const (
	AnyType    ValueType = ""
	IntType              = "int"
	UintType             = "uint"
	BoolType             = "bool"
	FloatType            = "float"
	BigIntType           = "bigint"
	StringType           = "string"
	BytesType            = "bytes"
	UUIDType             = "uuid"
	TupleType            = "tuple"
)

func AllTypes() []ValueType {
	return []ValueType{
		AnyType,
		IntType,
		UintType,
		BoolType,
		FloatType,
		BigIntType,
		StringType,
		BytesType,
		UUIDType,
		TupleType,
	}
}
