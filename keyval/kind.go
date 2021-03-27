package keyval

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
)

// IsVariable returns true if the KeyValue contains a Variable.
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

// DirIsVariable returns true if the Directory contains a Variable.
func DirIsVariable(dir Directory) bool {
	for _, element := range dir {
		if _, ok := element.(Variable); ok {
			return true
		}
	}
	return false
}

// TupIsVariable returns true if the Tuple contains a Variable.
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

// ValIsVariable returns true if the Value contains a Variable.
func ValIsVariable(val Value) bool {
	switch val.(type) {
	case Tuple:
		return TupIsVariable(val.(Tuple))
	case Variable:
		return true
	}
	return false
}
