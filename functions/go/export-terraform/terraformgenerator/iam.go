// Copyright 2022 Google LLC
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

package terraformgenerator

type iamBinding struct {
	Member   string
	Role     string
	resource *terraformResource
}

func (resource *terraformResource) HasIAMBindings() bool {
	bindings := resource.GetIAMBindings()
	return len(bindings) > 0
}

// Retrieve IAM bindings for a given resource
func (resource *terraformResource) GetIAMBindings() map[string][]*iamBinding {
	bindings := make(map[string][]*iamBinding)

	for _, child := range resource.Children {
		switch child.Kind {
		case "IAMPolicyMember":
			binding := &iamBinding{
				Member:   child.GetStringFromObject("spec", "member"),
				Role:     child.GetStringFromObject("spec", "role"),
				resource: child,
			}
			bindings[binding.Role] = append(bindings[binding.Role], binding)
		}
	}

	return bindings
}
