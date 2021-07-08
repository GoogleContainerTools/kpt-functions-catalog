package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/put-setter-values/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/put-setter-values/putsettervalues"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	psvp := PutSetterValuesProcessor{}
	cmd := command.Build(&psvp, command.StandaloneEnabled, false)

	cmd.Short = generated.PutSetterValuesShort
	cmd.Long = generated.PutSetterValuesLong
	cmd.Example = generated.PutSetterValuesExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type PutSetterValuesProcessor struct{}

func (asp *PutSetterValuesProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "put-setter-values",
	}
	items, err := run(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func run(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
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
func getSetters(fc *kyaml.RNode) (putsettervalues.PutSetterValues, error) {
	var fcd putsettervalues.PutSetterValues
	putsettervalues.Decode(fc, &fcd)
	return fcd, nil
}

// resultsToItems converts the results to
// equivalent items([]framework.Item)
func resultsToItems(sr putsettervalues.PutSetterValues) ([]framework.ResultItem, error) {
	var items []framework.ResultItem
	if len(sr.Results) == 0 {
		return nil, fmt.Errorf(`"apply-setters" function declaration not found, please add "apply-setters" function to the list of mutators in the Kptfile and retry`)
	}
	for _, res := range sr.Results {
		items = append(items, framework.ResultItem{
			Message: fmt.Sprintf("put setter value %q for setter with name %q", res.Value, res.Name),
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to put setter values: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
