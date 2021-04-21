package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"k8s.io/client-go/util/jsonpath"
)

var data = map[string]interface{}{
	"a": map[string]interface{}{
		"b": []interface{}{
			"c",
			"d",
		},
	},
}

func jpfn(expr string, data interface{}) (interface{}, error) {
	jp := jsonpath.New("jp")
	if err := jp.Parse(expr); err != nil {
		return nil, fmt.Errorf("unrecognized column definition %q", expr)
	}
	jp.AllowMissingKeys(true)
	// jp.EnableJSONOutput(true)

	var buf bytes.Buffer
	err := jp.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	var v interface{}
	err = json.Unmarshal(buf.Bytes(), &v)
	return v, err
}

func main() {
	fm := sprig.TxtFuncMap()
	fm["jp"] = jpfn

	tpl := template.Must(template.New("").Funcs(fm).Parse(`
{{ jp "{.a.b}" . | len }}
`))
	err := tpl.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
