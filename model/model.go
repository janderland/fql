package model

type (
	Query struct {
		Key   Key
		Value Value
	}

	Key struct {
		Directory Directory
		Tuple     Tuple
	}

	Directory []string

	// Contains nil, bool, string,
	// int64, uint64, float64,
	// tup.UUID, or Variable.
	Tuple []interface{}

	// Contains nil, bool, string,
	// int64, uint64, float64,
	// tup.UUID, Tuple, or
	// Variable.
	Value interface{}

	Variable struct {
		Name string
	}

	Clear struct{}
)
