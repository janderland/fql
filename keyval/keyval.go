// Package keyval provides the key value data structure
// and functions for inspecting the key values.
package keyval

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

	// A Tuple is equivalent to an FDB tuple. This type may
	// contain nil, bool, string, int64, float64, UUID, or
	// Variable.
	Tuple []interface{}

	// A Value represents an FDB value stored alongside
	// a key. This type may contain nil, bool, string,
	// uint64, float64, UUID, Tuple, Variable, or Clear.
	Value interface{}

	// A UUID is equivalent to an FDB UUID.
	UUID [16]byte

	// A Variable is used as a placeholder for any valid
	// values within the contexts that allow it.
	Variable struct {
		Name string
	}

	// A Clear is a special kind of Value which designates
	// a KeyValue as a clear operation.
	Clear struct{}
)
