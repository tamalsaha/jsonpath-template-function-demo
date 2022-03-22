package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/Masterminds/sprig/v3"
	"gomodules.xyz/jsonpath"
	core "k8s.io/api/core/v1"
	metatable "k8s.io/apimachinery/pkg/api/meta/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

// https://kubernetes.io/docs/reference/kubectl/jsonpath/

// kuubectl get cm kube-root-ca.crt -o json
var d3 = `{
    "apiVersion": "v1",
    "data": {
        "ca.crt": "-----BEGIN CERTIFICATE-----\nMIIC/jCCAeagAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl\ncm5ldGVzMB4XDTIyMDIxNDA1MzQxM1oXDTMyMDIxMjA1MzQxM1owFTETMBEGA1UE\nAxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL9n\nwMyAS74A+5Fx2yXcBzdIOxUkF+jD5jRoQp0qZYS65ROYYAzTVmF1IhJeSz8gUsSa\nMisKbnGfHFiTmX2K4UDzop/45CKBGYinW3JrxF2uIhu4+K21bU0l6/dRA4ACFQIF\nRA1uI6t3nYwfPEdzY4QiSwswQWso/Ev/jXu1dteaMNc2YDZSSK2QTs3jNjMFhuVc\nx0MR85ybRZykVOh2Yj4e4050DmuR776mVqulMuJRj8aGv6RpTAxVmTXSOGT7nAty\nDUvCXVyF/uCeFYXtf2UrDo2ZaQIQaLlcO0XHkwYLSgjHdEPCLTiXH4k1Ux/cfOya\n92FHvCKx+UT+o1OZD+cCAwEAAaNZMFcwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB\n/wQFMAMBAf8wHQYDVR0OBBYEFNCqwatFOSjh3PyRHAEujeceSVvWMBUGA1UdEQQO\nMAyCCmt1YmVybmV0ZXMwDQYJKoZIhvcNAQELBQADggEBAFT4zYlgvTZ12wRXHEL1\nNN3zjSDpxUuMJs/nLAegQKkOi69LIuh8yifCJTZhXAz9ZCfnFNkuOPiYm90iNLqa\n/n2XZ7mW6p+z5MWN/Ff9o3WJzBTEnkm8N3JMHkScb57I702QD7KBTCUEaTP1NAQq\n7N7NF9drEpBRlfuVhanfrgz/0CD0sc2cdNjE8xU7j/Yh54i5rg4NebqTHUB8stf2\nK7I177dU9RfZ0tCx+6ditQ6SY6woCwMFXKDLq8MFuk2AJZYTBVELd1lMLokJlXQW\nFxSuvpUVjKcdUlMx4wgd1qsuRbyihKZxqMiAp2F7X44VXreQrYkaVceT1x4zugjz\nscY=\n-----END CERTIFICATE-----\n"
    },
    "kind": "ConfigMap",
    "metadata": {
        "annotations": {
            "kubernetes.io/description": "Contains a CA bundle that can be used to verify the kube-apiserver when using internal endpoints such as the internal service IP or kubernetes.default.svc. No other usage is guaranteed across distributions of Kubernetes clusters."
        },
        "creationTimestamp": "2022-02-14T05:34:47Z",
        "managedFields": [
            {
                "apiVersion": "v1",
                "fieldsType": "FieldsV1",
                "fieldsV1": {
                    "f:data": {
                        ".": {},
                        "f:ca.crt": {}
                    },
                    "f:metadata": {
                        "f:annotations": {
                            ".": {},
                            "f:kubernetes.io/description": {}
                        }
                    }
                },
                "manager": "kube-controller-manager",
                "operation": "Update",
                "time": "2022-02-14T05:34:47Z"
            }
        ],
        "name": "kube-root-ca.crt",
        "namespace": "default",
        "resourceVersion": "440",
        "uid": "671c3f08-774d-454d-b2e2-fa03b049bd97"
    }
}`

var data = `{
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
        "creationTimestamp": "2021-12-20T02:36:40Z",
        "labels": {
            "component": "apiserver",
            "provider": "kubernetes"
        },
        "managedFields": [
            {
                "apiVersion": "v1",
                "fieldsType": "FieldsV1",
                "fieldsV1": {
                    "f:metadata": {
                        "f:labels": {
                            ".": {},
                            "f:component": {},
                            "f:provider": {}
                        }
                    },
                    "f:spec": {
                        "f:clusterIP": {},
                        "f:internalTrafficPolicy": {},
                        "f:ipFamilyPolicy": {},
                        "f:ports": {
                            ".": {},
                            "k:{\"port\":443,\"protocol\":\"TCP\"}": {
                                ".": {},
                                "f:name": {},
                                "f:port": {},
                                "f:protocol": {},
                                "f:targetPort": {}
                            }
                        },
                        "f:sessionAffinity": {},
                        "f:type": {}
                    }
                },
                "manager": "kube-apiserver",
                "operation": "Update",
                "time": "2021-12-20T02:36:40Z"
            }
        ],
        "name": "kubernetes",
        "namespace": "default",
        "resourceVersion": "209",
        "uid": "2329aab6-e918-4c74-9b23-884d2f016281"
    },
    "spec": {
        "clusterIP": "10.96.0.1",
        "clusterIPs": [
            "10.96.0.1"
        ],
        "internalTrafficPolicy": "Cluster",
        "ipFamilies": [
            "IPv4"
        ],
        "ipFamilyPolicy": "SingleStack",
        "ports": [
            {
                "name": "https",
                "port": 443,
                "protocol": "TCP",
                "targetPort": 6443
            }
        ],
        "sessionAffinity": "None",
        "type": "ClusterIP"
    },
    "status": {
        "loadBalancer": {}
    }
}`

var data3 = `{
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

func servicePortsFn(data interface{}) (string, error) {
	var ports []core.ServicePort

	if s, ok := data.(string); ok && s != "" {
		err := json.Unmarshal([]byte(s), &ports)
		if err != nil {
			return "", err
		}
	} else if _, ok := data.([]interface{}); ok {
		data, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(data, &ports)
		if err != nil {
			return "", err
		}
	}
	return MakeServicePortString(ports), nil
}

func MakeServicePortString(ports []core.ServicePort) string {
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
func main_tpl() {
	var d interface{}
	err := json.Unmarshal([]byte(d3), &d)
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
	fm["k8s_svc_ports"] = servicePortsFn

	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .spec.ports | k8s_svc_ports }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .metadata.abc.xyz | k8s_age }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .spec.template.spec.command | fmt_list }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .metadata.labels | fmt_labels }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .spec.selector2 | k8s_selector }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ printf "%s/%s" .metadata.namespace2 .metadata.name }}`))
	// tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ .metadata.namespace2 }}/{{ .metadata.namespace2 }}`))
	// Not that zero will attempt to add default values for types it knows,
	// but will still emit <no value> for others. We mitigate that later.

	tpl := template.Must(template.New("").Funcs(fm).Parse(`{{ dig "data" "ca.crt" "unknown" . }}`))
	tpl.Option("missingkey=default")
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

var doc5 = `{
  "dashboards": [
    {
      "title": "a"
    },
    {
      "title": "b"
    }
  ]
}`

func main() {
	var d interface{}
	err := json.Unmarshal([]byte(doc5), &d)
	if err != nil {
		panic(err)
	}

	enableJSONoutput := false
	expr := "{.dashboards[0]}"

	jp := jsonpath.New("")
	if err := jp.Parse(expr); err != nil {
		panic(err)
	}
	jp.AllowMissingKeys(false)
	jp.EnableJSONOutput(enableJSONoutput)

	var buf bytes.Buffer
	err = jp.Execute(&buf, d)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())

	//if enableJSONoutput {
	//	var v []interface{}
	//	err = json.Unmarshal(buf.Bytes(), &v)
	//	return v, err
	//}
	//return buf.String(), err
}
