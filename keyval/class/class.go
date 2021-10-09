package class

import q "github.com/janderland/fdbq/keyval"

// Class categorizes a KeyValue.
type Class string

const (
	// Constant specifies that the KeyValue has no Variable,
	// MaybeMore, or Clear. This kind of KeyValue can be used to
	// perform a set operation or is returned by a get operation.
	Constant Class = "constant"

	// Clear specifies that the KeyValue has no Variable or
	// MaybeMore and has a Clear Value. This kind of KeyValue can
	// be used to perform a clear operation.
	Clear Class = "clear"

	// SingleRead specifies that the KeyValue has a Variable
	// Value and doesn't have a Variable or MaybeMore in its Key.
	// This kind of KeyValue can be used to perform a get operation
	// that returns a single KeyValue.
	SingleRead Class = "single"

	// RangeRead specifies that the KeyValue has a Variable
	// or MaybeMore in its Key and doesn't have a Clear Value.
	// This kind of KeyValue can be used to perform a get
	// operation that returns multiple KeyValue.
	RangeRead Class = "range"

	// VariableClear specifies that the KeyValue has a
	// Variable or MaybeMore in its Key and has a Clear for
	// its value. This is an invalid class of KeyValue.
	VariableClear Class = "variable clear"
)

// subClass categorizes the Key, Directory,
// Tuple, and Value within a KeyValue.
type subClass int

const (
	// constantSubClass specifies that the component contains no
	// Variable, MaybeMore, or Clear.
	constantSubClass subClass = iota

	// variableSubClass specifies that the component contains a
	// Variable or MaybeMore.
	variableSubClass

	// clearSubClass specifies that the component contains a Clear.
	clearSubClass
)

// Classify returns the Class of the given KeyValue.
func Classify(kv q.KeyValue) Class {
	switch classifyKey(kv.Key) {
	case constantSubClass:
		switch classifyValue(kv.Value) {
		case clearSubClass:
			return Clear
		case variableSubClass:
			return SingleRead
		default:
			return Constant
		}
	default:
		switch classifyValue(kv.Value) {
		case clearSubClass:
			return VariableClear
		default:
			return RangeRead
		}
	}
}

func classifyKey(key q.Key) subClass {
	if classifyDir(key.Directory) == variableSubClass {
		return variableSubClass
	}
	if classifyTuple(key.Tuple) == variableSubClass {
		return variableSubClass
	}
	return constantSubClass
}

func classifyDir(dir q.Directory) subClass {
	class := dirClassification{}
	for _, element := range dir {
		element.DirElement(&class)
	}
	return class.result
}

func classifyTuple(tup q.Tuple) subClass {
	class := tupClassification{}
	for _, element := range tup {
		element.TupElement(&class)
	}
	return class.result
}

func classifyValue(val q.Value) subClass {
	class := valClassification{}
	val.Value(&class)
	return class.result
}
