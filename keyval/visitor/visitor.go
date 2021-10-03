package main

import (
	"strings"
	"text/template"

	g "github.com/janderland/fdbq/internal/generate"
)

func main() {
	var gen visitorGen
	g.Generate(&gen, []g.Input{
		{Type: g.Flag, Dst: &gen.visitor, Key: "visitor"},
		{Type: g.Flag, Dst: &gen.acceptor, Key: "acceptor"},
		{Type: g.Flag, Dst: &gen.types, Key: "types"},
	})
}

type visitorGen struct {
	visitor  string
	acceptor string
	types    string
}

func (x visitorGen) Name() string {
	return x.visitor
}

func (x visitorGen) Data() interface{} {
	return x
}

func (x visitorGen) Template() *template.Template {
	return template.Must(template.New("").Parse(`
type (
	{{.Visitor}} interface {
		{{range $i, $type := .Types -}}
		{{$.VisitorMethod $type}}({{$type}})
		{{end}}
	}

	{{.Acceptor}} interface {
		{{.AcceptorMethod}}({{.Visitor}})
	}
)

func _() {
	var (
		{{range .Types -}}
		{{.}} {{.}}
		{{end}}

		{{range .Types -}}
		_ {{$.Acceptor}} = &{{.}}
		{{end}}
	)
}

{{range $i, $type := .Types}}
func (x {{$type}}) {{$.AcceptorMethod}}(v {{$.Visitor}}) {
	v.{{$.VisitorMethod $type}}(x)
}
{{end}}
`))
}

func (x visitorGen) Visitor() string {
	return x.visitor + "Visitor"
}

func (x visitorGen) Acceptor() string {
	return x.acceptor
}

func (x visitorGen) Types() []string {
	return strings.Split(x.types, ",")
}

func (x visitorGen) VisitorMethod(typ string) string {
	return "Visit" + strings.Title(typ)
}

func (x visitorGen) AcceptorMethod() string {
	return strings.Title(x.acceptor)
}
