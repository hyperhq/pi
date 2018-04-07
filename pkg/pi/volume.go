/*
Copyright 2015 The Kubernetes Authors.

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
	"strconv"

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"

	"k8s.io/apimachinery/pkg/runtime"
)

// VolumeGeneratorV1 supports stable generation of an opaque volume
type VolumeGeneratorV1 struct {
	// Name of volume (required)
	Name string
	// Size of volume (optional)
	Size string
	// Zone of the volume (optional)
	Zone string
}

// Generate returns a volume using the specified parameters
func (s VolumeGeneratorV1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	err := ValidateParams(s.ParamNames(), genericParams)
	if err != nil {
		return nil, err
	}
	delegate := &VolumeGeneratorV1{}
	params := map[string]string{}
	for key, value := range genericParams {
		strVal, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("expected string, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}
	delegate.Name = params["name"]
	delegate.Size = params["size"]
	delegate.Zone = params["zone"]

	return delegate.StructuredGenerate()
}

// ParamNames returns the set of supported input parameters when using the parameter injection generator pattern
func (s VolumeGeneratorV1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"name", true},
		{"size", false},
		{"zone", false},
	}
}

// StructuredGenerate outputs a volume object using the configured fields
func (s VolumeGeneratorV1) StructuredGenerate() (runtime.Object, error) {
	var err error
	if err = s.validate(); err != nil {
		return nil, err
	}
	volume := &hyper.VolumeCreateRequest{}
	volume.Name = s.Name
	if s.Size != "" {
		if volume.Size, err = strconv.Atoi(s.Size); err != nil {
			return nil, err
		}
	}
	if s.Zone != "" {
		volume.Zone = s.Zone
	}
	return volume, nil
}

// validate validates required fields are set to support structured generation
func (s VolumeGeneratorV1) validate() error {
	if len(s.Name) == 0 {
		return fmt.Errorf("name must be specified")
	}
	return nil
}
