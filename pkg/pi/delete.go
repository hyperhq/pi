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

package pi

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	coreclient "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
)

const (
	Interval = time.Second * 1
	Timeout  = time.Minute * 5
)

// A Reaper terminates an object as gracefully as possible.
type Reaper interface {
	// Stop a given object within a namespace. timeout is how long we'll
	// wait for the termination to be successful. gracePeriod is time given
	// to an API object for it to delete itself cleanly (e.g., pod
	// shutdown). It may or may not be supported by the API object.
	Stop(namespace, name string, timeout time.Duration, gracePeriod *metav1.DeleteOptions) error
}

type NoSuchReaperError struct {
	kind schema.GroupKind
}

func (n *NoSuchReaperError) Error() string {
	return fmt.Sprintf("no reaper has been implemented for %v", n.kind)
}

func IsNoSuchReaperError(err error) bool {
	_, ok := err.(*NoSuchReaperError)
	return ok
}

func ReaperFor(kind schema.GroupKind, c internalclientset.Interface) (Reaper, error) {
	switch kind {
	case api.Kind("Pod"):
		return &PodReaper{c.Core()}, nil
	}
	return nil, &NoSuchReaperError{kind}
}

type PodReaper struct {
	client coreclient.PodsGetter
}

func (reaper *PodReaper) Stop(namespace, name string, timeout time.Duration, gracePeriod *metav1.DeleteOptions) error {
	pods := reaper.client.Pods(namespace)
	_, err := pods.Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return pods.Delete(name, gracePeriod)
}
