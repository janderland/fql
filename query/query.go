package query

type (
	Query struct {
		Key   Key
		Value Value
	}

	Key struct {
		Directory Directory
		Tuple     Tuple
	}

	// Contains string or Variable.
	Directory []interface{}

	// Contains nil, bool, string,
	// int64, uint64, float64,
	// UUID, or Variable.
	Tuple []interface{}

	// Contains nil, bool, string,
	// int64, uint64, float64,
	// UUID, Tuple, or Variable.
	Value interface{}

	UUID [16]byte

	Variable struct {
		Name string
	}

	Clear struct{}
)
