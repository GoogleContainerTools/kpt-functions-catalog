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
		sr, err := getSearchParams(resourceList.FunctionConfig)
		if err != nil {
			return fmt.Errorf("failed to parse function config: %w", err)
		}
		_, err = sr.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to perform search operation: %w", err)
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
	return `` //TODO: will add it in the next PR
}

// getSearchParams retrieve the search parameters from input config
func getSearchParams(fc interface{}) (SearchReplace, error) {
	var fcd SearchReplace
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fcd, fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fcd, fmt.Errorf("failed to parse input from function config: %w", err)
	}

	decode(rn, &fcd)
	return fcd, nil
}

// decode decodes the input yaml node into SearchReplace struct
func decode(rn *kyaml.RNode, fcd *SearchReplace) {
	dm := rn.GetDataMap()
	fcd.ByPath = getValue(dm, "by-path")
	fcd.ByValue = getValue(dm, "by-value")
	fcd.ByValueRegex = getValue(dm, "by-value-regex")
	fcd.PutValue = getValue(dm, "put-value")
	fcd.PutComment = getValue(dm, "put-comment")
}

// getValue returns the value for 'key' in map 'm'
// returns empty string if 'key' doesn't exist in 'm'
func getValue(m map[string]string, key string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return ""
}
