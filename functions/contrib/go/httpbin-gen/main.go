// Copyright 2020 Google LLC
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

// main.go
package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	resourceList := &framework.ResourceList{}
	cmd := framework.Command(resourceList, func() error {
		// cmd.Execute() will parse the ResourceList.functionConfig into cmd.Flags from
		// the ResourceList.functionConfig.data field.

		deploymentExists := false
		nrReplicas := 4
		newDep := yaml.MustParse(fmt.Sprintf(httpbinDeployment, nrReplicas))
		for i := range resourceList.Items {
			resource := resourceList.Items[i]
			m, _ := resource.GetMeta()
			if m.APIVersion != "apps/v1" {
				continue
			}
			resource, err = resource.Pipe(yaml.Lookup(
				"spec",
				"template",
				"spec",
				"containers",
				"[name=httpbin]"))
			if err == nil && resource != nil {
				// found the httpbin, update the http deployment
				deploymentExists = true
				// TODO(droot): update just the replica field
				resourceList.Items[i] = newDep
			}
		}
		if !deploymentExists {
			resourceList.Items = append(resourceList.Items, newDep)
		}
		return nil
	})
	// cmd.SetOut(ioutil.Discard)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

const httpbinDeployment = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: httpbin
spec:
  replicas: %d
  selector:
    matchLabels:
      app: httpbin
  template:
    metadata:
      labels:
        app: httpbin
    spec:
      containers:
      - name: httpbin
        image: kennethreitz/httpbin
        ports:
        - containerPort: 9876
`
