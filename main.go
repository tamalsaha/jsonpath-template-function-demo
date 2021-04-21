package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gomodules.xyz/jsonpath"
	core "k8s.io/api/core/v1"
	metatable "k8s.io/apimachinery/pkg/api/meta/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// https://kubernetes.io/docs/reference/kubectl/jsonpath/

var data2 = `{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "deployment.kubernetes.io/revision": "1"
    },
    "creationTimestamp": "2021-04-21T11:46:25Z",
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

var data = `{
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

func selectorFn(data string) (string, error) {
	var sel metav1.LabelSelector
	err := json.Unmarshal([]byte(data), &sel)
	if err != nil {
		return "", err
	}
	return metav1.FormatLabelSelector(&sel), nil
}

func ageFn(data string) (string, error) {
	var timestamp metav1.Time
	err := timestamp.UnmarshalQueryParameter(data)
	if err != nil {
		return "", err
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

	tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ jp "{.spec.ports}" . | k8s_ports }}`))
	err = tpl.Execute(os.Stdout, d)
	if err != nil {
		panic(err)
	}
}
