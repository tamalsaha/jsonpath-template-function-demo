package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gomodules.xyz/jsonpath"
	core "k8s.io/api/core/v1"
	metatable "k8s.io/apimachinery/pkg/api/meta/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

// https://kubernetes.io/docs/reference/kubectl/jsonpath/

var data = `{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "deployment.kubernetes.io/revision": "1"
    },
    "creationTimestamp": null,
    "labels": {
      "app": "busy-dep"
    },
    "name": "busy-dep",
    "namespace": "default"
  },
  "spec": {
    "progressDeadlineSeconds": 600,
    "replicas": 1,
    "revisionHistoryLimit": 10,
    "selector": {
      "matchLabels": {
        "app": "busy-dep"
      }
    },
    "strategy": {
      "rollingUpdate": {
        "maxSurge": "25%",
        "maxUnavailable": "25%"
      },
      "type": "RollingUpdate"
    },
    "template": {
      "metadata": {
        "creationTimestamp": null,
        "labels": {
          "app": "busy-dep"
        }
      },
      "spec": {
		"command": [
		  "sleep",
		  "3600"
		],
        "containers": [
          {
            "command": [
              "sleep",
              "3600"
            ],
            "image": "busybox",
            "imagePullPolicy": "IfNotPresent",
            "name": "busybox",
            "terminationMessagePath": "/dev/termination-log",
            "terminationMessagePolicy": "File"
          },
          {
            "command": [
              "sleep",
              "3600"
            ],
            "image": "ubuntu:18.04",
            "imagePullPolicy": "IfNotPresent",
            "name": "ubuntu",
            "terminationMessagePath": "/dev/termination-log",
            "terminationMessagePolicy": "File"
          }
        ],
        "dnsPolicy": "ClusterFirst",
        "restartPolicy": "Always",
        "schedulerName": "default-scheduler",
        "terminationGracePeriodSeconds": 30
      }
    }
  }
}`

var data2 = `{
  "apiVersion": "v1",
  "kind": "Service",
  "metadata": {
    "labels": {
      "component": "apiserver",
      "provider": "kubernetes"
    },
    "name": "kubernetes",
    "namespace": "default"
  },
  "spec": {
    "clusterIP": "10.96.0.1",
    "clusterIPs": [
      "10.96.0.1"
    ],
    "ports": [
      {
        "name": "https",
        "port": 443,
        "protocol": "TCP",
        "targetPort": 6443
      }
    ]
  }
}`

func jpfn(expr string, data interface{}, jsonoutput ...bool) (interface{}, error) {
	enableJSONoutput := len(jsonoutput) > 0 && jsonoutput[0]

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

func selectorFn(data interface{}) (string, error) {
	var sel metav1.LabelSelector
	if s, ok := data.(string); ok && s != "" {
		err := json.Unmarshal([]byte(s), &sel)
		if err != nil {
			return "", err
		}
	} else if _, ok := data.(map[string]interface{}); ok {
		err := meta_util.DecodeObject(data, &sel)
		if err != nil {
			return "", err
		}
	}
	return metav1.FormatLabelSelector(&sel), nil
}

func ageFn(data interface{}) (string, error) {
	var timestamp metav1.Time
	if s, ok := data.(string); ok && s != "" {
		err := timestamp.UnmarshalQueryParameter(s)
		if err != nil {
			return "", err
		}
	} else if _, ok := data.(map[string]interface{}); ok {
		err := meta_util.DecodeObject(data, &timestamp)
		if err != nil {
			return "", err
		}
	}
	return metatable.ConvertToHumanReadableDateType(timestamp), nil
}

func portsFn(data string) (string, error) {
	var ports []core.ServicePort
	err := json.Unmarshal([]byte(data), &ports)
	if err != nil {
		return "", err
	}
	return MakePortString(ports), nil
}

func formatLabelsFn(data interface{}) (string, error) {
	var lbl map[string]string
	if s, ok := data.(string); ok && s != "" {
		err := json.Unmarshal([]byte(s), &lbl)
		if err != nil {
			return "", err
		}
	} else if _, ok := data.(map[string]interface{}); ok {
		err := meta_util.DecodeObject(data, &lbl)
		if err != nil {
			return "", err
		}
	}
	return labels.FormatLabels(lbl), nil
}

func MakePortString(ports []core.ServicePort) string {
	pieces := make([]string, len(ports))
	for ix := range ports {
		port := &ports[ix]
		pieces[ix] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		if port.NodePort > 0 {
			pieces[ix] = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		}
	}
	return strings.Join(pieces, ",")
}

func main2(d interface{}) error {
	path := jsonpath.New("jp")
	expr := "{.metadata.labels}"
	if err := path.Parse(expr); err != nil {
		return fmt.Errorf("unrecognized column definition %q", expr)
	}
	path.AllowMissingKeys(true)

	results, err := path.FindResults(d)
	if err != nil {
		return err
	}

	x := results[0][0].Interface()
	fmt.Println(x)

	return nil
}

func fmtListFn(data interface{}) (string, error) {
	if s, ok := data.(string); ok && s != "" {
		return s, nil
	} else if arr, ok := data.([]interface{}); ok {
		s, err := json.Marshal(arr)
		return string(s), err
	}
	return "[]", nil
}

// "2021-04-21T11:46:25Z"
func main() {
	var d interface{}
	err := json.Unmarshal([]byte(data), &d)
	if err != nil {
		panic(err)
	}

	//err = main2(d)
	//if err != nil {
	//	panic(err)
	//}

	fm := sprig.TxtFuncMap()
	fm["jp"] = jpfn
	fm["k8s_selector"] = selectorFn
	fm["k8s_age"] = ageFn
	fm["k8s_ports"] = portsFn
	fm["fmt_labels"] = formatLabelsFn
	fm["fmt_list"] = fmtListFn

	tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .metadata.creationTimestamp | k8s_age }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .spec.template.spec.command | fmt_list }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .metadata.labels | fmt_labels }}`))
	//tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .spec.selector2 | k8s_selector }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ printf "%s/%s" .metadata.namespace2 .metadata.name }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .metadata.namespace2 }}/{{ .metadata.namespace2 }}`))
	// Not that zero will attempt to add default values for types it knows,
	// but will still emit <no value> for others. We mitigate that later.
	tpl.Option("missingkey=zero")
	err = tpl.Execute(os.Stdout, d)
	if err != nil {
		panic(err)
	}
}

func main______() {
	var m map[string]interface{}
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	var x map[string]interface{}
	err = json.Unmarshal(data, &x)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", x)
}

func main____() {
	type Person struct {
		Name string `json:"name,omitempty"`
	}

	data := []interface{}{
		map[string]interface{}{
			"name": "x",
		},
		map[string]interface{}{
			"name": "y",
		},
	}

	var persons []Person
	err := meta_util.DecodeObject(data, &persons)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", persons)
}
