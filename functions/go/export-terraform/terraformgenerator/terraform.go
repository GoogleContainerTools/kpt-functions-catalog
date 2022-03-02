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
	"strings"
	"text/template"
)

type TerraformFile struct {
	name string
}

func (rs *terraformResources) getHCL() (map[string]string, error) {
	files := []*TerraformFile{
		{
			name: "folders.tf",
		},
		{
			name: "iam.tf",
		},
	}

	tmpl, err := template.New("").ParseFS(templates, "templates/*")
	if err != nil {
		return nil, err
	}

	groupedResources := rs.getGrouped()

	// fmt.Printf("resources: %v\n", groupedResources)

	data := make(map[string]string)
	for _, file := range files {
		err := addFile(tmpl, file, groupedResources, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func addFile(tmpl *template.Template, file *TerraformFile, groupedResources map[string][]*terraformResource, data map[string]string) error {
	builder := strings.Builder{}
	wr := &(builder)

	err := tmpl.ExecuteTemplate(wr, file.name, groupedResources)
	if err != nil {
		return nil
	}

	content := strings.TrimSpace(builder.String())

	if len(content) < 1 {
		return nil
	}

	content = content + "\n"
	data[file.name] = content

	return nil
}
