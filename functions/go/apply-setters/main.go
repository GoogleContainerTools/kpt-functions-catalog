package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-setters/applysetters"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-setters/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		resourceList.Result = &framework.Result{
			Name: "apply-setters",
		}
		s, err := getSetters(resourceList.FunctionConfig)
		if err != nil {
			return fmt.Errorf("failed to parse function config: %w", err)
		}
		_, err = s.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to apply setters: %w", err)
		}
		resourceList.Result.Items = resultsToItems(s)
		return nil
	})

	cmd.Short = generated.ApplySettersShort
	cmd.Long = generated.ApplySettersLong
	cmd.Example = generated.ApplySettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getSetters retrieve the setters from input config
func getSetters(fc interface{}) (applysetters.ApplySetters, error) {
	var fcd applysetters.ApplySetters
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fcd, fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fcd, fmt.Errorf("failed to parse input from function config: %w", err)
	}
	applysetters.Decode(rn, &fcd)
	return fcd, nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func resultsToItems(sr applysetters.ApplySetters) []framework.Item {
	var items []framework.Item
	if len(sr.Results) == 0 {
		items = append(items, framework.Item{
			Message: "no matches",
		})
		return items
	}
	for _, res := range sr.Results {
		items = append(items, framework.Item{
			Message: fmt.Sprintf("set field value to %q", res.Value),
			Field:   framework.Field{Path: res.FieldPath},
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items
}
