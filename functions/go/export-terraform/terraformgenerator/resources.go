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

import (
	"fmt"
	"regexp"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
)

type terraformResource struct {
	Name     string
	Kind     string
	Item     *sdk.KubeObject
	Parent   *terraformResource
	Children []*terraformResource
	isChild  bool
}

func getDisplayName(ref *terraformResource) string {
	var displayName string
	found, err := ref.Item.Get(&displayName, "spec", "displayName")
	if err != nil || !found {
		// TODO: log failure to find
		displayName = ref.Item.Name()
	}
	return displayName
}

var tfNameRegex = regexp.MustCompile(`[^a-zA-Z\d_-]`)

func getResourceName(ref *terraformResource) string {
	name := ref.Item.Name()
	if name != "" {
		name = tfNameRegex.ReplaceAllString(name, "-")
		return name
	}
	return "organization"
}

func getTerraformId(ref *terraformResource) string {
	if ref.Kind == "Organization" {
		return fmt.Sprintf("\"organizations/%s\"", ref.Name)
	}
	return fmt.Sprintf("google_folder.%s.name", getResourceName(ref))
}
