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

func (rs *terraformResources) getHCL() (map[string]string, error) {
	tmpl, err := template.New("").ParseFS(templates, "templates/*")
	if err != nil {
		return nil, err
	}

	groupedResources := rs.getGrouped()

	// fmt.Printf("resources: %v\n", groupedResources)

	data := make(map[string]string)
	resourceFiles := []string{"folders.tf", "iam.tf", "projects.tf"}
	for _, file := range resourceFiles {
		err := addFile(tmpl, file, groupedResources, data)
		if err != nil {
			return nil, err
		}
	}

	// only add versions.tf if other files exist
	if len(data) > 0 {
		err = addFile(tmpl, "versions.tf", rs, data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func addFile(tmpl *template.Template, name string, inputData interface{}, data map[string]string) error {
	builder := strings.Builder{}
	wr := &(builder)

	err := tmpl.ExecuteTemplate(wr, name, inputData)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(builder.String())

	if len(content) < 1 {
		return nil
	}

	content = content + "\n"
	data[name] = content

	return nil
}
