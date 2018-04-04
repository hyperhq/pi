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

// A set of common functions needed by cmd/pi and pkg/pi packages.
package pi

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NamespaceInfo struct {
	Namespace string
}

// ResourceShortcuts represents a structure that holds the information how to
// transition from resource's shortcut to its full name.
type ResourceShortcuts struct {
	ShortForm schema.GroupResource
	LongForm  schema.GroupResource
}

// ResourcesShortcutStatic is the list of short names to their expanded names.
// Note that the list is ordered by group.
var ResourcesShortcutStatic = []ResourceShortcuts{
	// If you add an entry here, please also take a look at pkg/pi/cmd/cmd.go
	// and add an entry to valid_resources when appropriate.
	{
		ShortForm: schema.GroupResource{Resource: "no"},
		LongForm:  schema.GroupResource{Resource: "nodes"},
	},
	{
		ShortForm: schema.GroupResource{Resource: "po"},
		LongForm:  schema.GroupResource{Resource: "pods"},
	},
	{
		ShortForm: schema.GroupResource{Resource: "svc"},
		LongForm:  schema.GroupResource{Resource: "services"},
	},
}

// ResourceShortFormFor looks up for a short form of resource names.
// TODO: Change the signature of this function so that it can
// make use of ResourceShortcuts.
func ResourceShortFormFor(resource string) (string, bool) {
	var alias string
	exists := false
	for _, item := range ResourcesShortcutStatic {
		if item.LongForm.Resource == resource {
			alias = item.ShortForm.Resource
			exists = true
			break
		}
	}
	return alias, exists
}

// ResourceAliases returns the resource shortcuts and plural forms for the given resources.
func ResourceAliases(rs []string) []string {
	as := make([]string, 0, len(rs))
	plurals := make(map[string]struct{}, len(rs))
	for _, r := range rs {
		var plural string
		switch {
		case r == "endpoints":
			plural = r // exception. "endpoint" does not exist. Why?
		case strings.HasSuffix(r, "y"):
			plural = r[0:len(r)-1] + "ies"
		case strings.HasSuffix(r, "s"):
			plural = r + "es"
		default:
			plural = r + "s"
		}
		as = append(as, plural)

		plurals[plural] = struct{}{}
	}

	for _, item := range ResourcesShortcutStatic {
		if _, found := plurals[item.LongForm.Resource]; found {
			as = append(as, item.ShortForm.Resource)
		}
	}
	return as
}
