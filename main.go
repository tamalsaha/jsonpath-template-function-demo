package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gomodules.xyz/jsonpath"
)

// https://kubernetes.io/docs/reference/kubectl/jsonpath/

var data = `{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "deployment.kubernetes.io/revision": "1"
    },
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

	tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ jp "{.metadata.labels}" . }}`))
	err = tpl.Execute(os.Stdout, d)
	if err != nil {
		panic(err)
	}
}
