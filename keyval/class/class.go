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

type character struct {
	vstampFutures int
	hasVariable   bool
	hasClear      bool
	hasNil        bool
}

func (x *character) orFields(c character) character {
	return character{
		vstampFutures: x.vstampFutures + c.vstampFutures,
		hasVariable:   x.hasVariable || c.hasVariable,
		hasClear:      x.hasClear || c.hasClear,
		hasNil:        x.hasNil || c.hasNil,
	}
}

func (x *character) String() string {
	var (
		str   strings.Builder
		empty = true
	)
	str.WriteRune('[')
	if x.hasVariable {
		empty = false
		str.WriteString("var")
	}
	if x.vstampFutures > 0 {
		if !empty {
			str.WriteRune(',')
		}
		empty = false
		str.WriteString("vstamps")
	}
	if x.hasClear {
		if !empty {
			str.WriteRune(',')
		}
		empty = false
		str.WriteString("clear")
	}
	if x.hasNil {
		if !empty {
			str.WriteRune(',')
		}
		str.WriteString("nil")
	}
	str.WriteRune(']')
	return str.String()
}

// invalidClass returns a Class describing the
// invalid characterstics of the key-value.
func invalidClass(c character) Class {
	return Class(fmt.Sprintf("invalid%s", c.String()))
}

// Classify returns the Class of the given KeyValue.
func Classify(kv q.KeyValue) Class {
	dirClass := classifyDir(kv.Key.Directory)
	keyClass := dirClass.orFields(classifyTuple(kv.Key.Tuple))
	kvClass := keyClass.orFields(classifyValue(kv.Value))

	// KeyValues should never contain `nil`.
	if kvClass.hasNil {
		return invalidClass(kvClass)
	}

	// KeyValues should contain, at most, 1 VStampFuture.
	if kvClass.vstampFutures > 1 {
		return invalidClass(kvClass)
	}

	// Ensure that, at most, one of the conditions are true.
	bools := []bool{kvClass.vstampFutures > 0, kvClass.hasVariable, kvClass.hasClear}
	for i, b1 := range bools {
		for j, b2 := range bools {
			if i == j {
				continue
			}
			if b1 && b2 {
				return invalidClass(kvClass)
			}
		}
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

func classifyDir(dir q.Directory) character {
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
	return class.orFields(character{hasNil: hasNil})
}

func classifyTuple(tup q.Tuple) character {
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
	return class.orFields(character{hasNil: hasNil})
}

func classifyValue(val q.Value) character {
	var class valClassification
	if val == nil {
		return character{hasNil: true}
	}
	val.Value(&class)
	return class.orFields(character{})
}
