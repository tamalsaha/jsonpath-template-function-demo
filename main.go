package main

import (
	"os"
	"text/template"
	"github.com/Masterminds/sprig/v3"
	"k8s.io/client-go/util/jsonpath"
)

func main2() {
	jp := jsonpath.New(col.Name)
	if err := jp.Parse(col.JSONPath); err != nil {
		return nil, fmt.Errorf("unrecognized column definition %q", col.JSONPath)
	}
	jp.AllowMissingKeys(true)

}

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