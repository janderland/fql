package keyval

type (
	KeyValue struct {
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

func IsVariable(kv *KeyValue) bool {
	if DirIsVariable(kv.Key.Directory) {
		return true
	}
	if TupIsVariable(kv.Key.Tuple) {
		return true
	}
	if ValIsVariable(kv.Value) {
		return true
	}
	return false
}

func DirIsVariable(dir Directory) bool {
	for _, element := range dir {
		if _, ok := element.(Variable); ok {
			return true
		}
	}
	return false
}

func TupIsVariable(tup Tuple) bool {
	for _, element := range tup {
		switch element.(type) {
		case Tuple:
			if TupIsVariable(element.(Tuple)) {
				return true
			}
		case Variable:
			return true
		}
	}
	return false
}

func ValIsVariable(val Value) bool {
	switch val.(type) {
	case Tuple:
		if TupIsVariable(val.(Tuple)) {
			return true
		}
	case Variable:
		return true
	}
	return false
}
