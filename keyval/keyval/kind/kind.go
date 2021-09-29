package kind

import (
	q "github.com/janderland/fdbq/keyval/keyval"
)

// Kind categorizes a KeyValue.
type Kind string

const (
	// ConstantKind specifies that the KeyValue has no Variable,
	// MaybeMore, or Clear. This kind of KeyValue can be used to
	// perform a set operation or is returned by a get operation.
	ConstantKind Kind = "constant"

	// ClearKind specifies that the KeyValue has no Variable or
	// MaybeMore and has a Clear Value. This kind of KeyValue can
	// be used to perform a clear operation.
	ClearKind Kind = "clear"

	// SingleReadKind specifies that the KeyValue has a Variable
	// Value and doesn't have a Variable or MaybeMore in its Key.
	// This kind of KeyValue can be used to perform a get operation
	// that returns a single KeyValue.
	SingleReadKind Kind = "single"

	// RangeReadKind specifies that the KeyValue has a Variable
	// or MaybeMore in its Key and doesn't have a Clear Value.
	// This kind of KeyValue can be used to perform a get
	// operation that returns multiple KeyValue.
	RangeReadKind Kind = "range"

	// VariableClearKind specifies that the KeyValue has a
	// Variable or MaybeMore in its Key and has a Clear for
	// its value. This is an invalid kind of KeyValue.
	VariableClearKind Kind = "variable clear"
)

// subKind categorizes the Key, Directory,
// Tuple, and Value within a KeyValue.
type subKind = int

const (
	// constantSubKind specifies that the component contains no
	// Variable, MaybeMore, or Clear.
	constantSubKind subKind = iota

	// variableSubKind specifies that the component contains a
	// Variable or MaybeMore.
	variableSubKind

	// clearSubKind specifies that the component contains a Clear.
	clearSubKind
)

// Which returns the Kind of the given KeyValue. If the KeyValue
// is malformed then InvalidKind and a non-nil error are returned.
// For details on what a malformed KeyValue is, see the KeyValue,
// Key, Directory, Tuple, and Value documentation.
func Which(kv q.KeyValue) Kind {
	switch keyKind(kv.Key) {
	case constantSubKind:
		switch valKind(kv.Value) {
		case clearSubKind:
			return ClearKind
		case variableSubKind:
			return SingleReadKind
		default:
			return ConstantKind
		}
	default:
		switch valKind(kv.Value) {
		case clearSubKind:
			return VariableClearKind
		default:
			return RangeReadKind
		}
	}
}

func keyKind(key q.Key) subKind {
	if dirKind(key.Directory) == variableSubKind {
		return variableSubKind
	}
	if tupKind(key.Tuple) == variableSubKind {
		return variableSubKind
	}
	return constantSubKind
}

func dirKind(dir q.Directory) subKind {
	v := dirVisitor{kind: constantSubKind}
	for _, e := range dir {
		e.DirElement(&v)
	}
	return v.kind
}

func tupKind(tup q.Tuple) subKind {
	v := tupVisitor{kind: constantSubKind}
	for _, e := range tup {
		e.TupElement(&v)
	}
	return v.kind
}

func valKind(val q.Value) subKind {
	v := valVisitor{kind: constantSubKind}
	val.Value(&v)
	return v.kind
}
