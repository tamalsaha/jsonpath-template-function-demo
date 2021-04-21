package main

import (
	"os"
	"text/template"
	"github.com/Masterminds/sprig/v3"
)

func main() {
	type Inner struct {
		A string
	}
	type Outer struct {
		Inner
	}

	type NA struct {
		O []Outer
	}

	na := NA{
		O: []Outer{
			{
				Inner: Inner{A: "123"},
			},
		},
	}
	tpl := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(`
{{ define "t2" }}
{{ printf "%v" .A }}
{{ end }}
{{ range $svc := .O }}
	{{ template "t2" $svc }}
{{ end }}
`))
	tpl.Execute(os.Stdout, &na)
}