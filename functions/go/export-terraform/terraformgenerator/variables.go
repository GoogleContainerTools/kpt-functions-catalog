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

package terraformgenerator

type variable struct {
	Name        string
	Description string
	Default     string
}

// Find common values and make them into variables
func (rs *terraformResources) makeVariables() {
	rs.Variables = make(map[string]*variable)

	resources := rs.getGrouped()

	// If we have a single org, make it a variable
	if len(resources["Organization"]) == 1 {
		org := resources["Organization"][0]
		orgVar := &variable{
			Name:        "org_id",
			Description: "The organization id for the associated resources",
			Default:     org.Name,
		}
		rs.Variables["org_id"] = orgVar
		org.variable = orgVar
	}
}
