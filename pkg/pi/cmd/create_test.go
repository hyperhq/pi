/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/hyperhq/client-go/dynamic"
	"github.com/hyperhq/client-go/rest/fake"
	cmdtesting "github.com/hyperhq/pi/pkg/pi/cmd/testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

var unstructuredSerializer = dynamic.ContentConfig().NegotiatedSerializer

func TestExtraArgsFail(t *testing.T) {
	initTestErrorHandler(t)
	buf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	f, _, _, _ := cmdtesting.NewAPIFactory()
	c := NewCmdCreate(f, buf, errBuf)
	options := CreateOptions{}
	if options.ValidateArgs(c, []string{"rc"}) == nil {
		t.Errorf("unexpected non-error")
	}
}

func TestCreateObject(t *testing.T) {
	initTestErrorHandler(t)
	pods, _ := testData()
	pods.Items[0].Name = "nginx"

	f, tf, codec, _ := cmdtesting.NewAPIFactory()
	tf.Printer = &testPrinter{}
	tf.UnstructuredClient = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/namespaces/default/pods" && m == http.MethodPost:
				return &http.Response{StatusCode: http.StatusCreated, Header: defaultHeader(), Body: objBody(codec, &pods.Items[0])}, nil
			default:
				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	tf.Namespace = "default"
	buf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	cmd := NewCmdCreate(f, buf, errBuf)
	cmd.Flags().Set("filename", "../../../examples/pod/pod-nginx.yaml")
	cmd.Flags().Set("output", "name")
	cmd.Run(cmd, []string{})

	// uses the name from the file, not the response
	if buf.String() != "pod/nginx\n" {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestCreateMultipleObject(t *testing.T) {
	initTestErrorHandler(t)
	pods, svcs := testData()

	f, tf, codec, _ := cmdtesting.NewAPIFactory()
	tf.Printer = &testPrinter{}
	tf.UnstructuredClient = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/namespaces/default/services" && m == http.MethodPost:
				return &http.Response{StatusCode: http.StatusCreated, Header: defaultHeader(), Body: objBody(codec, &svcs.Items[0])}, nil
			case p == "/namespaces/default/pods" && m == http.MethodPost:
				return &http.Response{StatusCode: http.StatusCreated, Header: defaultHeader(), Body: objBody(codec, &pods.Items[0])}, nil
			default:
				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	tf.Namespace = "default"
	buf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	cmd := NewCmdCreate(f, buf, errBuf)
	cmd.Flags().Set("filename", "../../../examples/pod/pod-nginx.yaml")
	cmd.Flags().Set("filename", "../../../examples/service/service-clusterip-default.yaml")
	cmd.Flags().Set("output", "name")
	cmd.Run(cmd, []string{})

	// Names should come from the REST response, NOT the files
	if buf.String() != "replicationcontroller/rc1\nservice/baz\n" {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestCreateDirectory(t *testing.T) {
	initTestErrorHandler(t)
	pods, _ := testData()
	pods.Items[0].Name = "name"

	f, tf, codec, _ := cmdtesting.NewAPIFactory()
	tf.Printer = &testPrinter{}
	tf.UnstructuredClient = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: "v1"},
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/namespaces/default/pods" && m == http.MethodPost:
				return &http.Response{StatusCode: http.StatusCreated, Header: defaultHeader(), Body: objBody(codec, &pods.Items[0])}, nil
			default:
				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	tf.Namespace = "default"
	buf := bytes.NewBuffer([]byte{})
	errBuf := bytes.NewBuffer([]byte{})

	cmd := NewCmdCreate(f, buf, errBuf)
	cmd.Flags().Set("filename", "../../../examples/pod/pod-nginx.yaml")
	cmd.Flags().Set("output", "name")
	cmd.Run(cmd, []string{})

	if buf.String() != "pod/name\npod/name\npod/name\n" {
		t.Errorf("unexpected output: %s", buf.String())
	}
}
