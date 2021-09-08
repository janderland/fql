package main

import (
	"strings"
	"text/template"

	g "github.com/janderland/fdbq/internal/generate"
)

func main() {
	var gen VisitorGen
	g.Generate(&gen, []g.Input{
		{Type: g.Flag, Dst: &gen.visitor, Key: "visitor"},
		{Type: g.Flag, Dst: &gen.acceptor, Key: "acceptor"},
		{Type: g.Flag, Dst: &gen.types, Key: "types"},
	})
}

type VisitorGen struct {
	visitor  string
	acceptor string
	types    string
}

func (x VisitorGen) Name() string {
	return x.visitor
}

func (x VisitorGen) Data() interface{} {
	return x
}

func (x VisitorGen) Template() *template.Template {
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
		{{.}}_ {{.}}
		{{end}}

		{{range .Types -}}
		_ {{$.Acceptor}} = &{{.}}_
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

func (x VisitorGen) Visitor() string {
	return x.visitor + "Visitor"
}

func (x VisitorGen) Acceptor() string {
	return x.acceptor
}

func (x VisitorGen) Types() []string {
	return strings.Split(x.types, ",")
}

func (x VisitorGen) VisitorMethod(typ string) string {
	return "Visit" + strings.Title(typ)
}

func (x VisitorGen) AcceptorMethod() string {
	return strings.Title(x.acceptor)
}
