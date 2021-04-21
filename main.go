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

// https://kubernetes.io/docs/reference/kubectl/jsonpath/

var data = `{
  "kind": "List",
  "items":[
    {
      "kind":"None",
      "metadata":{"name":"127.0.0.1"},
      "status":{
        "capacity":{"cpu":"4"},
        "addresses":[{"type": "LegacyHostIP", "address":"127.0.0.1"}]
      }
    },
    {
      "kind":"None",
      "metadata":{"name":"127.0.0.2"},
      "status":{
        "capacity":{"cpu":"8"},
        "addresses":[
          {"type": "LegacyHostIP", "address":"127.0.0.2"},
          {"type": "another", "address":"127.0.0.3"}
        ]
      }
    }
  ],
  "users":[
    {
      "name": "myself",
      "user": {}
    },
    {
      "name": "e2e",
      "user": {"username": "admin", "password": "secret"}
    }
  ]
}`

func jpfn(enableJSONoutput bool) func(_ string, _ interface{}) (interface{}, error) {
	return func(expr string, data interface{}) (interface{}, error) {
		jp := jsonpath.New("jp")
		if err := jp.Parse(expr); err != nil {
			return nil, fmt.Errorf("unrecognized column definition %q", expr)
		}
		jp.AllowMissingKeys(true)
		jp.EnableJSONOutput(enableJSONoutput)

		var buf bytes.Buffer
		err := jp.Execute(&buf, data)
		if err != nil {
			return nil, err
		}

		if enableJSONoutput {
			var v []interface{}
			err = json.Unmarshal(buf.Bytes(), &v)
			return v, err
		}
		return buf.String(), err
	}
}

func main() {
	var d interface{}
	err := json.Unmarshal([]byte(data), &d)
	if err != nil {
		panic(err)
	}

	fm := sprig.TxtFuncMap()
	fm["j"] = jpfn(true)
	fm["js"] = jpfn(false)

	tpl := template.Must(template.New("").Funcs(fm).Parse(`
{{ js "{.items}" . | fromJson | len }}
`))
	err = tpl.Execute(os.Stdout, d)
	if err != nil {
		panic(err)
	}
}
