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
	// MaybeMore, Clear, or VStampFuture. This kind of KeyValue
	// can be used to perform a set operation or is returned by
	// a get operation.
	Constant Class = "constant"

	// VStamp specifies that the KeyValue of the Constant class, but
	// it contains a VStampFuture. This kind of KeyValue can only be
	// used to perform a set operation. When the KeyValue is later
	// read, the VStampFuture we be replaced by a VStamp.
	VStamp Class = "vstamp"

	// Clear specifies that the KeyValue has no Variable, MaybeMore,
	// or VStampFuture and has Clear as it's value. This kind of
	// KeyValue can be used to perform a clear operation.
	Clear Class = "clear"

	// ReadSingle specifies that the KeyValue has a Variable as it's
	// value and doesn't have a Variable or MaybeMore in its Key. It
	// also must not contain a VStampFuture anywhere in the KeyValue.
	// This kind of KeyValue can be used to perform a get operation
	// that returns a single KeyValue.
	ReadSingle Class = "single"

	// ReadRange specifies that the KeyValue has a Variable
	// or MaybeMore in its key and doesn't have a Clear as it's
	// value. This kind of KeyValue can be used to perform a get
	// operation that returns multiple KeyValue.
	ReadRange Class = "range"
)

// Classify returns the Class of the given KeyValue.
func Classify(kv q.KeyValue) Class {
	dirAttr := getAttributesOfDir(kv.Key.Directory)
	keyAttr := dirAttr.merge(getAttributesOfTup(kv.Key.Tuple))
	kvAttr := keyAttr.merge(getAttributesOfVal(kv.Value))

	// KeyValues should never contain `nil`.
	if kvAttr.hasNil {
		return invalidClass(kvAttr)
	}

	// KeyValues should contain, at most, 1 VStampFuture.
	if kvAttr.vstampFutures > 1 {
		return invalidClass(kvAttr)
	}

	// Ensure that, at most, one of these conditions are true.
	count := 0
	for _, cond := range []bool{
		kvAttr.vstampFutures > 0,
		kvAttr.hasVariable,
		kvAttr.hasClear,
	} {
		if cond {
			count++
		}
	}
	if count > 1 {
		return invalidClass(kvAttr)
	}

	switch {
	case keyAttr.hasVariable:
		return ReadRange
	case kvAttr.hasVariable:
		return ReadSingle
	case kvAttr.vstampFutures > 0:
		return VStamp
	case kvAttr.hasClear:
		return Clear
	default:
		return Constant
	}
}

// attributes describes the characterstics of a KeyValue,
// which are relevant to it's classification.
type attributes struct {
	vstampFutures int
	hasVariable   bool
	hasClear      bool
	hasNil        bool
}

// merge combines the attributes of parts of a KeyValue
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
// invalid KeyValue. The relevant attributes are
// included in the Class to assist with debugging.
func invalidClass(attr attributes) Class {
	var (
		str   strings.Builder
		empty = true
	)
	for substr, cond := range map[string]bool{
		fmt.Sprintf("vstamps:%d", attr.vstampFutures): attr.vstampFutures > 0,
		"var":   attr.hasVariable,
		"clear": attr.hasClear,
		"nil":   attr.hasNil,
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
