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

// HasVariable returns true if the KeyValue contains a Variable.
func HasVariable(kv *KeyValue) bool {
	if DirHasVariable(kv.Key.Directory) {
		return true
	}
	if TupHasVariable(kv.Key.Tuple) {
		return true
	}
	if ValHasVariable(kv.Value) {
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

// TupHasVariable returns true if the Tuple contains a Variable.
func TupHasVariable(tup Tuple) bool {
	for _, element := range tup {
		switch element.(type) {
		case Tuple:
			if TupHasVariable(element.(Tuple)) {
				return true
			}
		case Variable:
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
