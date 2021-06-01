// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gatekeeper/generated"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	k8syaml "sigs.k8s.io/yaml"
)

func main() {
	resourceList := &framework.ResourceList{}

	cmd := framework.Command(resourceList, func() error {
		var objects []runtime.Object
		for _, item := range resourceList.Items {
			meta, err := item.GetValidatedMetadata()
			if err != nil {
				return err
			}

			s, err := item.String()
			if err != nil {
				return err
			}
			obj, err := scheme.New(schema.FromAPIVersionAndKind(meta.APIVersion, meta.Kind))
			switch {
			case runtime.IsNotRegisteredError(err):
				obj = &unstructured.Unstructured{}
			case err != nil:
				return err
			}
			err = k8syaml.Unmarshal([]byte(s), obj)
			if err != nil {
				return err
			}
			objects = append(objects, obj)
		}

		result, err := Validate(objects)
		// When err is not nil, result should be nil.
		if err != nil {
			result = &framework.Result{
				Items: []framework.Item{
					{
						Message:  err.Error(),
						Severity: framework.Error,
					},
				},
			}
		}
		resourceList.Result = result
		if resultContainsError(result) {
			return result
		}
		return nil
	})
	cmd.Short = generated.GatekeeperShort
	cmd.Long = generated.GatekeeperLong
	cmd.Example = generated.GatekeeperExamples
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func resultContainsError(result *framework.Result) bool {
	if result == nil {
		return false
	}
	for _, item := range result.Items {
		if item.Severity == framework.Error {
			return true
		}
	}
	return false
}
