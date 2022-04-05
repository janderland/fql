package format

import (
	"strings"

	q "github.com/janderland/fdbq/keyval"
	"github.com/janderland/fdbq/parser/parser/internal"
)

func Query(qry q.Query) string {
	var fmt query
	qry.Query(&fmt)
	return fmt.str
}

func Directory(dir q.Directory) string {
	var b strings.Builder
	var fmt directory

	for _, part := range dir {
		b.WriteRune(internal.DirSep)
		part.DirElement(&fmt)
		b.WriteString(fmt.str)
	}

	return b.String()
}

func Tuple(tup q.Tuple) string {
	var b strings.Builder
	var fmt tuple

	b.WriteRune(internal.TupStart)
	for i, element := range tup {
		if i != 0 {
			b.WriteRune(internal.TupSep)
		}
		element.TupElement(&fmt)
		b.WriteString(fmt.str)
	}
	b.WriteRune(internal.TupEnd)

	return b.String()
}

func Variable(v q.Variable) string {
	var b strings.Builder

	b.WriteRune(internal.VarStart)
	for i, vType := range v {
		if i != 0 {
			b.WriteRune(internal.VarSep)
		}
		b.WriteString(string(vType))
	}
	b.WriteRune(internal.VarEnd)

	return b.String()
}
