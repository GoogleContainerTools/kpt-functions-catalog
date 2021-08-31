package main

import (
	"fmt"
	"os"
	"strings"

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
	var prunedCount = 0
	var prunedNames []string
	fileNames := "local resources not found"

	var res []*yaml.RNode
	for _, node := range resourceList.Items {
		if node.IsNilOrEmpty() {
			continue
		}
		// only add the resources which are not local configs
		if node.GetAnnotations()[filters.LocalConfigAnnotation] != "true" {
			res = append(res, node)
		} else {
			prunedCount++
			prunedNames = append(prunedNames, node.GetName())
		}
	}

	if prunedCount > 0 {
		fileNames = strings.Join(prunedNames, ", ")
	}

	resourceList.Items = res
	resultMessage := fmt.Sprintf("Resources Pruned: [Count: %d, Names: {%s}]", prunedCount, fileNames)

	items = append(items, framework.ResultItem{
		Message: resultMessage,
	})

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

type RemoveLocalConfigResourcesConfigProcessor struct{}
