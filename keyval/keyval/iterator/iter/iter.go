package main

import (
	"strings"
	"text/template"

	g "github.com/janderland/fdbq/internal/generate"
)

func main() {
	var gen mustGen
	g.Generate(&gen, []g.Input{
		{Type: g.Flag, Dst: &gen.types, Key: "types"},
	})
}

type mustGen struct {
	types string
}

func (x mustGen) Name() string {
	return ""
}

func (x mustGen) Data() interface{} {
	return x
}

func (x mustGen) Template() *template.Template {
	return template.Must(template.New("").Parse(`
import q "github.com/janderland/fdbq/keyval/keyval"

{{range $i, $type := .Types}}
// {{$type}} asserts the current element of the tuple is of type {{$type}}.
// If the type assertion fails a ConversionError is returned. If the
// iterator points beyond the end of the tuple, a ShortTupleError is
// returned. Otherwise, the element is returned and the iterator is
// pointed at the next element.
func (x *TupleIterator) {{$type}}() (out q.{{$type}}, err error) {
	if x.i >= len(x.t) {
		panic(ShortTupleError)
	}

	var ok bool
	if out, ok = x.t[x.i].(q.{{$type}}); !ok {
		err = ConversionError{
			InValue: x.t[x.i],
			OutType: out,
			Index:   x.i,
		}
		return
	}

	x.i++
	return
}

// Must{{$type}} does the same thing as {{$type}}, except it panics any
// errors instead of returning them. These errors will be recovered by
// the wrapping call to ReadTuple and returned by that function.
func (x *TupleIterator) Must{{$type}}() q.{{$type}} {
	val, err := x.{{$type}}()
	if err != nil {
		panic(err)
	}
	return val
}
{{end}}
`))
}

func (x mustGen) Types() []string {
	return strings.Split(x.types, ",")
}
