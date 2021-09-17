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
	"io"
	"sort"
	"strings"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/fn/sdk"
)

type terraformResources struct {
	resources map[string]*terraformResource
}

func (rs *terraformResources) getResourceRef(kind string, name string, item *sdk.KubeObject) *terraformResource {
	if rs.resources == nil {
		rs.resources = map[string]*terraformResource{}
	}

	key := fmt.Sprintf("%s/%s", kind, name)
	resourceRef, found := rs.resources[key]
	if !found {
		resourceRef = &terraformResource{
			Name: name,
			Kind: kind,
		}
		rs.resources[key] = resourceRef
	}
	if item != nil {
		resourceRef.Item = item
	}
	return resourceRef
}

func getParentRef(item *sdk.KubeObject) (string, string, error) {
	paths := [][]string{
		{"Folder", "spec", "folderRef", "name"},
		{"Organization", "spec", "organizationRef", "external"},
		{"Project", "metadata", "annotations", "cnrm.cloud.google.com/my-project"},
	}

	var name string
	for _, path := range paths {
		found, err := item.Get(&name, path[1:]...)
		if err != nil || !found {
			continue
		}

		return path[0], strings.TrimSpace(name), nil
	}

	return "", "", fmt.Errorf("no parent reference found for %s", item.Name())
}

func appendTerraform(wr io.Writer, myRef *terraformResource, tmpl *template.Template) error {
	if (myRef.Kind) == "Organization" {
		return nil
	}
	err := tmpl.Execute(wr, myRef)
	if err != nil {
		return err
	}
	return nil
}

type TerraformFile struct {
	name     string
	kinds    map[string]bool
	contents strings.Builder
}

func (rs *terraformResources) getHCL() (map[string]string, error) {
	files := []*TerraformFile{
		{
			name:     "folders.tf",
			kinds:    map[string]bool{"Folder": true},
			contents: strings.Builder{},
		},
	}

	tmpl, err := template.New("folder.tf").Funcs(template.FuncMap{
		"displayName":  getDisplayName,
		"resourceName": getResourceName,
		"reference":    getTerraformId,
	}).ParseFS(templates, "templates/folder.tf")

	if err != nil {
		return nil, err
	}

	// iterate over resources in a stable oder
	keys := make([]string, len(rs.resources))
	i := 0
	for k := range rs.resources {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, key := range keys {
		resource := rs.resources[key]
		if resource.Item == nil {
			continue
		}
		for _, file := range files {
			_, ok := file.kinds[resource.Kind]
			sdk.Logf("found matching template for %s/%s: %t\n", resource.Kind, resource.Name, ok)
			if ok {
				err := appendTerraform(&(file.contents), resource, tmpl)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	data := make(map[string]string)
	for _, file := range files {
		if file.contents.Len() > 0 {
			data[file.name] = file.contents.String()
		}
	}

	return data, nil
}
