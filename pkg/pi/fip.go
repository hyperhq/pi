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

	"github.com/hyperhq/client-go/tools/clientcmd/api/hyper"

	"k8s.io/apimachinery/pkg/runtime"
)

// FipGeneratorV1 supports stable generation of an opaque fip
type FipGeneratorV1 struct {
	// Count of fip (required)
	Count int
}

type FipNameV1 struct {
	Name string
}

// Generate returns a fip using the specified parameters
func (s FipGeneratorV1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	err := ValidateParams(s.ParamNames(), genericParams)
	if err != nil {
		return nil, err
	}
	delegate := &FipGeneratorV1{}
	params := map[string]int{}
	for key, value := range genericParams {
		strVal, isInt := value.(int)
		if !isInt {
			return nil, fmt.Errorf("expected int, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}
	delegate.Count = params["count"]

	return delegate.StructuredGenerate()
}

// ParamNames returns the set of supported input parameters when using the parameter injection generator pattern
func (s FipGeneratorV1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"count", true},
	}
}

// StructuredGenerate outputs a fip object using the configured fields
func (s FipGeneratorV1) StructuredGenerate() (runtime.Object, error) {
	var err error
	if err = s.validate(); err != nil {
		return nil, err
	}
	fip := &hyper.FipAllocateRequest{}
	fip.Count = s.Count
	return fip, nil
}

// validate validates required fields are set to support structured generation
func (s FipGeneratorV1) validate() error {
	if s.Count <= 0 {
		return fmt.Errorf("count must be greater than 0")
	}
	return nil
}

////////////////////////////////////////////////////////
// Generate returns a fip using the specified parameters
func (s FipNameV1) Generate(genericParams map[string]interface{}) (runtime.Object, error) {
	err := ValidateParams(s.ParamNames(), genericParams)
	if err != nil {
		return nil, err
	}
	delegate := &FipNameV1{}
	params := map[string]string{}
	for key, value := range genericParams {
		strVal, isString := value.(string)
		if !isString {
			return nil, fmt.Errorf("expected int, saw %v for '%s'", value, key)
		}
		params[key] = strVal
	}
	delegate.Name = params["name"]

	return delegate.StructuredGenerate()
}

// ParamNames returns the set of supported input parameters when using the parameter injection generator pattern
func (s FipNameV1) ParamNames() []GeneratorParam {
	return []GeneratorParam{
		{"name", true},
	}
}

// StructuredGenerate outputs a fip object using the configured fields
func (s FipNameV1) StructuredGenerate() (runtime.Object, error) {
	var err error
	if err = s.validate(); err != nil {
		return nil, err
	}
	fip := &hyper.FipRenameRequest{}
	fip.Name = s.Name
	return fip, nil
}

// validate validates required fields are set to support structured generation
func (s FipNameV1) validate() error {
	if len(s.Name) == 0 {
		return fmt.Errorf("name can not be empty")
	}
	return nil
}
