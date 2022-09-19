// Package format converts key-values into query strings.
package format

import (
	"encoding/hex"
	"strconv"
	"strings"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/internal"
)

type Cfg struct {
	// When set to false, byte strings are formatted
	// as their length instead of the actual string.
	PrintBytes bool
}

// Format provides methods which convert the types
// defined in keyval into strings. The methods with
// an input parameter format their input into a string
// and append the string to an internal buffer, which
// can be retrieved or cleared via the String and
// Reset methods.
type Format struct {
	Builder *strings.Builder
	Cfg     Cfg
}

// String returns the contents of the internal buffer.
// TODO: Add escapes.
func (x *Format) String() string {
	return x.Builder.String()
}

// Reset clears the contents of the internal buffer.
func (x *Format) Reset() {
	x.Builder.Reset()
}

// Query formats the given keyval.Query and appends
// it to the internal buffer.
func (x *Format) Query(in q.Query) {
	in.Query(&formatQuery{x})
}

// KeyValue formats the given keyval.KeyValue
// and appends it to the internal buffer.
func (x *Format) KeyValue(in q.KeyValue) {
	x.Key(in.Key)
	x.Builder.WriteRune(internal.KeyValSep)
	x.Value(in.Value)
}

// Key formats the given keyval.Key
// and appends it to the internal buffer.
func (x *Format) Key(in q.Key) {
	x.Directory(in.Directory)
	x.Tuple(in.Tuple)
}

// Value formats the given keyval.Value
// and appends it to the internal buffer.
func (x *Format) Value(in q.Value) {
	in.Value(&formatData{x})
}

// Directory formats the given keyval.Directory
// and appends it to the internal buffer.
func (x *Format) Directory(in q.Directory) {
	for _, element := range in {
		x.Builder.WriteRune(internal.DirSep)
		element.DirElement(&formatDirElement{x})
	}
}

// Tuple formats the given keyval.Tuple
// and appends it to the internal buffer.
func (x *Format) Tuple(in q.Tuple) {
	x.Builder.WriteRune(internal.TupStart)
	for i, element := range in {
		if i != 0 {
			x.Builder.WriteRune(internal.TupSep)
		}
		element.TupElement(&formatData{x})
	}
	x.Builder.WriteRune(internal.TupEnd)
}

// Variable formats the given keyval.Variable
// and appends it to the internal buffer.
func (x *Format) Variable(in q.Variable) {
	x.Builder.WriteRune(internal.VarStart)
	for i, vType := range in {
		if i != 0 {
			x.Builder.WriteRune(internal.VarSep)
		}
		x.Builder.WriteString(string(vType))
	}
	x.Builder.WriteRune(internal.VarEnd)
}

// Bytes formats the given keyval.Bytes
// and appends it to the internal buffer.
func (x *Format) Bytes(in q.Bytes) {
	if x.Cfg.PrintBytes {
		x.Builder.WriteString(internal.HexStart)
		x.Builder.WriteString(hex.EncodeToString(in))
	} else {
		x.Builder.WriteString(strconv.FormatInt(int64(len(in)), 10))
		x.Builder.WriteString(" bytes")
	}
}

// Str formats the given keyval.String
// and appends it to the internal buffer.
func (x *Format) Str(in q.String) {
	x.Builder.WriteRune(internal.StrMark)
	x.Builder.WriteString(escapeString(string(in)))
	x.Builder.WriteRune(internal.StrMark)
}

// UUID formats the given keyval.UUID
// and appends it to the internal buffer.
func (x *Format) UUID(in q.UUID) {
	x.Builder.WriteString(hex.EncodeToString(in[:4]))
	x.Builder.WriteRune('-')
	x.Builder.WriteString(hex.EncodeToString(in[4:6]))
	x.Builder.WriteRune('-')
	x.Builder.WriteString(hex.EncodeToString(in[6:8]))
	x.Builder.WriteRune('-')
	x.Builder.WriteString(hex.EncodeToString(in[8:10]))
	x.Builder.WriteRune('-')
	x.Builder.WriteString(hex.EncodeToString(in[10:]))
}

// Bool formats the given keyval.Bool
// and appends it to the internal buffer.
func (x *Format) Bool(in q.Bool) {
	if in {
		x.Builder.WriteString(internal.True)
	} else {
		x.Builder.WriteString(internal.False)
	}
}

// Int formats the given keyval.Int
// and appends it to the internal buffer.
func (x *Format) Int(in q.Int) {
	x.Builder.WriteString(strconv.FormatInt(int64(in), 10))
}

// Uint formats the given keyval.Uint
// and appends it to the internal buffer.
func (x *Format) Uint(in q.Uint) {
	x.Builder.WriteString(strconv.FormatUint(uint64(in), 10))
}

// Float formats the given keyval.Float
// and appends it to the internal buffer.
func (x *Format) Float(in q.Float) {
	x.Builder.WriteString(strconv.FormatFloat(float64(in), 'g', 10, 64))
}

// Nil formats the given keyval.Nil
// and appends it to the internal buffer.
func (x *Format) Nil(_ q.Nil) {
	x.Builder.WriteString(internal.Nil)
}

// Clear formats the given keyval.Clear
// and appends it to the internal buffer.
func (x *Format) Clear(_ q.Clear) {
	x.Builder.WriteString(internal.Clear)
}

// MaybeMore formats the given keyval.MaybeMore
// and appends it to the internal buffer.
func (x *Format) MaybeMore(_ q.MaybeMore) {
	x.Builder.WriteString(internal.MaybeMore)
}

func escapeString(in string) string {
	out := strings.ReplaceAll(in, "\\", "\\\\")
	out = strings.ReplaceAll(out, "\"", "\\\"")
	return out
}
