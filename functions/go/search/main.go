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
			Name: "search",
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
	sr, err := getSearchParams(resourceList.FunctionConfig)
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
	return `Search and optionally replace fields across all resources.

Search matchers are provided with 'by-' prefix. When multiple matchers
are provided they are AND’ed together. 'put-' matchers are mutually exclusive.

Here are the list of matchers:

by-value
Match by value of a field.

by-value-regex
Match by Regex for the value of a field. The syntax of the regular expressions
accepted is the same general syntax used by Go, Perl, Python, and other languages.
More precisely, it is the syntax accepted by RE2 and described at
https://golang.org/s/re2syntax. With the exception that it matches the entire
value of the field by default without requiring start (^) and end ($) characters.

by-path
Match by path expression of a field. Path expressions are used to deeply navigate
and match particular yaml nodes. Please note that the path expressions are not
regular expressions.

put-value
Set or update the value of the matching fields. Input can be a pattern for which
the numbered capture groups are resolved using --by-value-regex input.

put-comment
Set or update the line comment for matching fields. Input can be a pattern for
which the numbered capture groups are resolved using --by-value-regex input.

To search and replace field value to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  by-value: nginx
  put-value: ubuntu

To put the setter pattern as a line comment for matching fields:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  by-value: my-project-id-foo
  put-value: 'kpt-set: ${project-id}-foo'

`
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
			Message:  fmt.Sprintf("failed to perform search operation: %q", errMsg),
			Severity: framework.Error,
		},
	}
}
