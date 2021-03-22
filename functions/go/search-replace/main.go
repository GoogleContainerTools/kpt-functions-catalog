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
		resourceList.Result = &framework.Result{
			Name: "search-replace",
		}
		items, err := run(resourceList)
		if err != nil {
			resourceList.Result.Items = getErrorItem(err.Error())
			return resourceList.Result
		}
		resourceList.Result.Items = items
		return nil
	})

	cmd.Long = usage()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// run resolves the function params and runs the function on resources
func run(resourceList *framework.ResourceList) ([]framework.Item, error) {
	sr, err := getSearchReplaceParams(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}

	_, err = sr.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}

	return searchResultsToItems(sr), nil
}
func usage() string {
	return `` // TODO: pmarupaka add usage docs
}

// getSearchReplaceParams retrieve the search parameters from input config
func getSearchReplaceParams(fc interface{}) (SearchReplace, error) {
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

// searchResultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func searchResultsToItems(sr SearchReplace) []framework.Item {
	var items []framework.Item
	for _, res := range sr.Results {

		var message string
		if sr.PutComment != "" || sr.PutValue != "" {
			message = fmt.Sprintf("Mutated field value to %q", res.Value)
		} else {
			message = fmt.Sprintf("Matched field value %q", res.Value)
		}

		items = append(items, framework.Item{
			Message: message,
			Field:   framework.Field{Path: res.FieldPath},
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.Item {
	return []framework.Item{
		{
			Message:  fmt.Sprintf("failed to perform search-replace operation: %q", errMsg),
			Severity: framework.Error,
		},
	}
}
