package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/list-setters/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/list-setters/listsetters"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

//nolint
func main() {
	lsp := ListSettersProcessor{}
	cmd := command.Build(&lsp, command.StandaloneEnabled, false)

	cmd.Short = generated.ListSettersShort
	cmd.Long = generated.ListSettersLong
	cmd.Example = generated.ListSettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ListSettersProcessor struct{}

func (lsp *ListSettersProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "list-setters",
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
	ls := listsetters.New()
	_, err := ls.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}
	resultItems, err := resultsToItems(ls)
	if err != nil {
		return nil, err
	}
	return resultItems, nil
}

// resultsToItems converts the listsetters results to
// equivalent items([]framework.Item)
func resultsToItems(sr listsetters.ListSetters) ([]framework.ResultItem, error) {
	var items []framework.ResultItem
	// add any warnings that occurred during setter discovery
	for _, w := range sr.Warnings {
		items = append(items, framework.ResultItem{
			Message:  w.Error(),
			Severity: framework.Warning,
		})
	}
	rs := sr.GetResults()
	if len(rs) == 0 {
		return nil, fmt.Errorf("no setters found")
	}
	for _, r := range rs {
		items = append(items, framework.ResultItem{
			Message: r.String(),
		})
	}
	return items, nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to list setters: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
