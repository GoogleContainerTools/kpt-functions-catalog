package applysetters

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type ApplySettersProcessor struct{}

func (asp *ApplySettersProcessor) Process(resourceList *framework.ResourceList) error {
	items, err := run(resourceList)
	if err != nil {
		resourceList.Results = []*framework.Result{getErrorItem(err.Error())}
		return err
	}
	resourceList.Results = items
	return nil
}

func run(resourceList *framework.ResourceList) ([]*framework.Result, error) {
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
func getSetters(fc *kyaml.RNode) (ApplySetters, error) {
	var fcd ApplySetters
	Decode(fc, &fcd)
	return fcd, nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func resultsToItems(sr ApplySetters) ([]*framework.Result, error) {
	var items []*framework.Result
	if len(sr.Results) == 0 {
		items = append(items, &framework.Result{
			Message:  "no matches for input setter(s)",
			Severity: framework.Warning,
		})
	}
	for _, res := range sr.Results {
		items = append(items, &framework.Result{
			Message:  fmt.Sprintf("set field value to %q", res.Value),
			Severity: framework.Info,
			Field:    &framework.Field{Path: res.FieldPath},
			File:     &framework.File{Path: res.FilePath},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) *framework.Result {
	return &framework.Result{
		Message:  fmt.Sprintf("failed to apply setters: %s", errMsg),
		Severity: framework.Error,
	}
}
