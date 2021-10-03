package class

import q "github.com/janderland/fdbq/keyval/keyval"

// Class categorizes a KeyValue.
type Class string

const (
	// ConstantClass specifies that the KeyValue has no Variable,
	// MaybeMore, or Clear. This kind of KeyValue can be used to
	// perform a set operation or is returned by a get operation.
	ConstantClass Class = "constant"

	// ClearClass specifies that the KeyValue has no Variable or
	// MaybeMore and has a Clear Value. This kind of KeyValue can
	// be used to perform a clear operation.
	ClearClass Class = "clear"

	// SingleReadClass specifies that the KeyValue has a Variable
	// Value and doesn't have a Variable or MaybeMore in its Key.
	// This kind of KeyValue can be used to perform a get operation
	// that returns a single KeyValue.
	SingleReadClass Class = "single"

	// RangeReadClass specifies that the KeyValue has a Variable
	// or MaybeMore in its Key and doesn't have a Clear Value.
	// This kind of KeyValue can be used to perform a get
	// operation that returns multiple KeyValue.
	RangeReadClass Class = "range"

	// VariableClearClass specifies that the KeyValue has a
	// Variable or MaybeMore in its Key and has a Clear for
	// its value. This is an invalid class of KeyValue.
	VariableClearClass Class = "variable clear"
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

// Which returns the Kind of the given KeyValue. If the KeyValue
// is malformed then InvalidKind and a non-nil error are returned.
// For details on what a malformed KeyValue is, see the KeyValue,
// Key, Directory, Tuple, and Value documentation.
func Which(kv q.KeyValue) Class {
	switch classifyKey(kv.Key) {
	case constantSubClass:
		switch classifyValue(kv.Value) {
		case clearSubClass:
			return ClearClass
		case variableSubClass:
			return SingleReadClass
		default:
			return ConstantClass
		}
	default:
		switch classifyValue(kv.Value) {
		case clearSubClass:
			return VariableClearClass
		default:
			return RangeReadClass
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
