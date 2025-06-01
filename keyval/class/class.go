// Package class classifies a key-value by the kind of operation it represents.
package class

import (
	"fmt"
	"strings"

	q "github.com/janderland/fql/keyval"
)

// Class categorizes a KeyValue.
type Class string

const (
	// Constant specifies that the KeyValue has no Variable,
	// MaybeMore, or Clear. This kind of KeyValue can be used to
	// perform a set operation or is returned by a get operation.
	Constant Class = "constant"

	// VStamp specifies that the KeyValue contains a VStampFuture.
	// This kind of KeyValue can only be used to perform a set
	// operation. When the KeyValue is later read, the VStampFuture
	// we be replaced by a VStamp.
	VStamp Class = "vstamp"

	// Clear specifies that the KeyValue has no Variable or
	// MaybeMore and has a Clear Value. This kind of KeyValue can
	// be used to perform a clear operation.
	Clear Class = "clear"

	// ReadSingle specifies that the KeyValue has a Variable
	// Value and doesn't have a Variable or MaybeMore in its Key.
	// This kind of KeyValue can be used to perform a get operation
	// that returns a single KeyValue.
	ReadSingle Class = "single"

	// ReadRange specifies that the KeyValue has a Variable
	// or MaybeMore in its Key and doesn't have a Clear Value.
	// This kind of KeyValue can be used to perform a get
	// operation that returns multiple KeyValue.
	ReadRange Class = "range"
)

// attributes describes the characterstics of a key-value,
// which are relevant to it's classification.
type attributes struct {
	vstampFutures int
	hasVariable   bool
	hasClear      bool
	hasNil        bool
}

// merge combines the attributes of parts of a key-value
// to infer the attributes of the whole key-value.
func (x *attributes) merge(c attributes) attributes {
	return attributes{
		vstampFutures: x.vstampFutures + c.vstampFutures,
		hasVariable:   x.hasVariable || c.hasVariable,
		hasClear:      x.hasClear || c.hasClear,
		hasNil:        x.hasNil || c.hasNil,
	}
}

// invalidClass creates a Class for a semantically
// invalid key-values. The attributes are included
// in the class to assist with debugging.
func invalidClass(attr attributes) Class {
	var (
		str   strings.Builder
		empty = true
	)
	for substr, cond := range map[string]bool{
		fmt.Sprintf("vstamps:%d", attr.vstampFutures): attr.vstampFutures > 0,
		"var": attr.hasVariable,
		"clear": attr.hasClear,
		"nil": attr.hasNil,
	} {
		if !cond {
			continue
		}
		if !empty {
			str.WriteRune(',')
		}
		str.WriteString(substr)
		empty = false
	}
	return Class(fmt.Sprintf("invalid[%s]", str.String()))
}

// Classify returns the Class of the given KeyValue.
func Classify(kv q.KeyValue) Class {
	dirClass := classifyDir(kv.Key.Directory)
	keyClass := dirClass.merge(classifyTuple(kv.Key.Tuple))
	kvClass := keyClass.merge(classifyValue(kv.Value))

	// KeyValues should never contain `nil`.
	if kvClass.hasNil {
		return invalidClass(kvClass)
	}

	// KeyValues should contain, at most, 1 VStampFuture.
	if kvClass.vstampFutures > 1 {
		return invalidClass(kvClass)
	}

	// Ensure that, at most, one of the conditions are true.
	count := 0
	for _, cond := range []bool{
		kvClass.vstampFutures > 0,
		kvClass.hasVariable,
		kvClass.hasClear,
	} {
		if cond {
			count++
		}
	}
	if count > 1 {
		return invalidClass(kvClass)
	}

	// After the loop above, we know at most
	// one of the cases below will execute.
	switch {
	case kvClass.hasVariable:
		if keyClass.hasVariable {
			return ReadRange
		}
		return ReadSingle
	case kvClass.vstampFutures > 0:
		return VStamp
	case kvClass.hasClear:
		return Clear
	default:
		return Constant
	}
}

func classifyDir(dir q.Directory) attributes {
	var (
		class  dirClassification
		hasNil bool
	)
	for _, element := range dir {
		if element == nil {
			hasNil = true
			continue
		}
		element.DirElement(&class)
	}
	return class.orFields(attributes{hasNil: hasNil})
}

func classifyTuple(tup q.Tuple) attributes {
	var (
		class  tupClassification
		hasNil bool
	)
	for _, element := range tup {
		if element == nil {
			hasNil = true
			continue
		}
		element.TupElement(&class)
	}
	return class.orFields(attributes{hasNil: hasNil})
}

func classifyValue(val q.Value) attributes {
	var class valClassification
	if val == nil {
		return attributes{hasNil: true}
	}
	val.Value(&class)
	return class.orFields(attributes{})
}
