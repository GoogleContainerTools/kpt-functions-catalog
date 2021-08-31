package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	fp := RemoveLocalConfigResourcesConfigProcessor{}
	cmd := command.Build(&fp, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func (fp *RemoveLocalConfigResourcesConfigProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "remove-local-config-resources",
	}

	items, err := ProcessResources(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func ProcessResources(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
	var items []framework.ResultItem
	results := []Result{}

	var res []*yaml.RNode
	for _, node := range resourceList.Items {
		if node.IsNilOrEmpty() {
			continue
		}
		// only add the resources which are not local configs
		if node.GetAnnotations()[filters.LocalConfigAnnotation] != "true" {
			res = append(res, node)
		} else {
			result := Result{Name: node.GetName()}
			results = append(results, result)
		}
	}

	resourceList.Items = res

	if len(results) > 0 {
		items = append(items, framework.ResultItem{
			Message: fmt.Sprintf("Number of resources pruned: %d", len(results)),
		})

		for _, result := range results {
			items = append(items, framework.ResultItem{
				Message: fmt.Sprintf("Resource name: [%s]", result.Name),
			})
		}
	} else if len(results) == 0 {
		item := framework.ResultItem{
			Message: "Found no resources to prune with the local config annotation",
		}

		item.Severity = framework.Warning
		items = append(items, item)
	}

	return items, nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to remove local configs: %s", errMsg),
			Severity: framework.Error,
		},
	}
}

type Result struct {
	Name string
}

type RemoveLocalConfigResourcesConfigProcessor struct{}
