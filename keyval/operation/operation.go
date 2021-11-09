package main

import (
	"strings"
	"text/template"

	g "github.com/janderland/fdbq/internal/generate"
)

func main() {
	var gen operationGen
	g.Generate(&gen, []g.Input{
		{Type: g.Flag, Dst: &gen.opName, Key: "op-name"},
		{Type: g.Flag, Dst: &gen.paramName, Key: "param-name"},
		{Type: g.Flag, Dst: &gen.types, Key: "types"},
	})
}

type operationGen struct {
	opName    string
	paramName string
	types     string
}

func (x operationGen) Name() string {
	return x.opName
}

func (x operationGen) Data() interface{} {
	return x
}

func (x operationGen) Template() *template.Template {
	return template.Must(template.New("").Parse(`
type (
	{{.OpName}} interface {
		{{range $i, $type := .Types -}}
		{{$.VisitorMethod $type}}({{$type}})
		{{end}}
	}

	{{.ParamName}} interface {
		{{.AcceptorMethod}}({{.OpName}})
		Eq(interface{}) bool
	}
)

func _() {
	var (
		{{range .Types -}}
		{{.}} {{.}}
		{{end}}

		{{range .Types -}}
		_ {{$.ParamName}} = &{{.}}
		{{end}}
	)
}

{{range $i, $type := .Types}}
func (x {{$type}}) {{$.AcceptorMethod}}(op {{$.OpName}}) {
	op.{{$.VisitorMethod $type}}(x)
}
{{end}}
`))
}

func (x operationGen) OpName() string {
	return x.opName + "Operation"
}

func (x operationGen) ParamName() string {
	return x.paramName
}

func (x operationGen) Types() []string {
	return strings.Split(x.types, ",")
}

func (x operationGen) VisitorMethod(typ string) string {
	return "For" + strings.Title(typ)
}

func (x operationGen) AcceptorMethod() string {
	return strings.Title(x.paramName)
}
