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
	return "must"
}

func (x mustGen) Data() interface{} {
	return x
}

func (x mustGen) Template() *template.Template {
	return template.Must(template.New("").Parse(`
import q "github.com/janderland/fdbq/keyval/keyval"

{{range $i, $type := .Types}}
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
