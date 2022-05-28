package format

import (
	"encoding/hex"
	"strconv"
	"strings"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/internal"
)

func Query(in q.Query) string {
	var op queryOp
	in.Query(&op)
	return op.str
}

func Keyval(in q.KeyValue) string {
	return Key(in.Key) + string(internal.KeyValSep) + Value(in.Value)
}

func Key(in q.Key) string {
	return Directory(in.Directory) + Tuple(in.Tuple)
}

func Value(in q.Value) string {
	var op valueOp
	in.Value(&op)
	return op.str
}

func Directory(in q.Directory) string {
	var b strings.Builder
	var op directoryOp

	for _, part := range in {
		b.WriteRune(internal.DirSep)
		part.DirElement(&op)
		b.WriteString(op.str)
	}

	return b.String()
}

func Tuple(in q.Tuple) string {
	var b strings.Builder
	var op tupleOp

	b.WriteRune(internal.TupStart)
	for i, element := range in {
		if i != 0 {
			b.WriteRune(internal.TupSep)
		}
		element.TupElement(&op)
		b.WriteString(op.str)
	}
	b.WriteRune(internal.TupEnd)

	return b.String()
}

func variable(in q.Variable) string {
	var b strings.Builder

	b.WriteRune(internal.VarStart)
	for i, vType := range in {
		if i != 0 {
			b.WriteRune(internal.VarSep)
		}
		b.WriteString(string(vType))
	}
	b.WriteRune(internal.VarEnd)

	return b.String()
}

func hexadecimal(in q.Bytes) string {
	var out strings.Builder

	out.WriteString(internal.HexStart)
	out.WriteString(hex.EncodeToString(in))

	return out.String()
}

func str(in q.String) string {
	var out strings.Builder

	out.WriteRune(internal.StrMark)
	out.WriteString(string(in))
	out.WriteRune(internal.StrMark)

	return out.String()
}

func uuid(in q.UUID) string {
	var out strings.Builder

	out.WriteString(hex.EncodeToString(in[:4]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[4:6]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[6:8]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[8:10]))
	out.WriteRune('-')
	out.WriteString(hex.EncodeToString(in[10:]))

	return out.String()
}

func boolean(in q.Bool) string {
	if in {
		return internal.True
	} else {
		return internal.False
	}
}

func integer(in q.Int) string {
	return strconv.FormatInt(int64(in), 10)
}

func unsigned(in q.Uint) string {
	return strconv.FormatUint(uint64(in), 10)
}

func float(in q.Float) string {
	return strconv.FormatFloat(float64(in), 'g', 10, 64)
}
