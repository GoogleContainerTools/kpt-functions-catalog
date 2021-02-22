package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		s, err := getSetters(resourceList.FunctionConfig)
		if err != nil {
			return fmt.Errorf("failed to parse function config: %w", err)
		}
		_, err = s.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to apply setters: %w", err)
		}
		return nil
	})

	cmd.Long = usage()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func usage() string {
	return `Apply setter values to resource fields with setter references.

Configured using a ConfigMap with key-value pairs in 'data' field in
'ConfigMap' resource. Example:

Example:

To apply a setter 'project_id: my-project' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  projectId: my-project

Values to array setters must be sequence nodes wrapped into strings

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  projectId: my-project
  environments: |
    - dev
    - staging
`
}

// getSetters retrieve the setters from input config
func getSetters(fc interface{}) (ApplySetters, error) {
	var fcd ApplySetters
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fcd, fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fcd, fmt.Errorf("failed to parse input from function config: %w", err)
	}

	return fcd, decode(rn, &fcd)
}

// decode decodes the input yaml node into Set struct
func decode(rn *kyaml.RNode, fcd *ApplySetters) error {
	for k, v := range rn.GetDataMap() {
		fcd.Setters = append(fcd.Setters, Setter{Name: k, Value: v})
	}
	return nil
}
