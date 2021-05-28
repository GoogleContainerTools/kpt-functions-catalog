package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/create-setters/createsetters"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/create-setters/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		resourceList.Result = &framework.Result{
			Name: "create-setters",
		}
		items, err := run(resourceList)
		if err != nil {
			resourceList.Result.Items = getErrorItem(err.Error())
			return resourceList.Result
		}
		resourceList.Result.Items = items
		return nil
	})

	cmd.Short = generated.CreateSettersShort
	cmd.Long = generated.CreateSettersLong
	cmd.Example = generated.CreateSettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(resourceList *framework.ResourceList) ([]framework.Item, error) {
	s, err := getSetters(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}
	_, err = s.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}
	resultItems, err := resultsToItems(s)
	if err != nil {
		return nil, err
	}
	return resultItems, nil
}

// getSetters retrieve the setters from input config
func getSetters(fc interface{}) (createsetters.CreateSetters, error) {
	var fcd createsetters.CreateSetters
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fcd, fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fcd, fmt.Errorf("failed to parse input from function config: %w", err)
	}
	createsetters.Decode(rn, &fcd)
	return fcd, nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func resultsToItems(cs createsetters.CreateSetters) ([]framework.Item, error) {
	var items []framework.Item
	if len(cs.Results) == 0 {
		return nil, fmt.Errorf("no matches for the input list of setters")
	}
	for _, res := range cs.Results {
		items = append(items, framework.Item{
			Message: fmt.Sprintf("set field value to %q", res.Value),
			Field:   framework.Field{Path: res.FieldPath},
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.Item {
	return []framework.Item{
		{
			Message:  fmt.Sprintf("failed to create setters: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
