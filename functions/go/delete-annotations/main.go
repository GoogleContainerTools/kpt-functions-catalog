package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	annotationKeysLiteral = "annotationKeys"
	annotationDelimeter   = ","
)

//nolint
func main() {
	fp := DeleteAnnotationsProcessor{}
	cmd := command.Build(&fp, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
func (fp *DeleteAnnotationsProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "delete-annotations",
	}

	// get the annotation key from functionConfig
	var annotationKeys string
	err := getAnnotationKeys(resourceList.FunctionConfig, &annotationKeys)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}

	// process resources in the package and delete the annotation keys provided
	items, err := processResources(resourceList.Items, annotationKeys)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func processResources(resourceList []*yaml.RNode, keys string) ([]framework.ResultItem, error) {
	var resultItems []framework.ResultItem
	for _, node := range resourceList {
		if node.IsNilOrEmpty() {
			continue
		}

		//confirm resources are valid
		metadata, err := node.GetMeta()
		if err != nil {
			return nil, err
		}

		if metadata.Name == "" || metadata.Kind == "" {
			continue
		}

		keysList := strings.Split(strings.TrimSpace(keys), annotationDelimeter)

		for _, annotationKey := range keysList {
			mutatedNode, err := node.Pipe(yaml.Lookup("metadata", "annotations"), yaml.FieldClearer{
				Name: annotationKey})

			if err != nil {
				return nil, err
			}

			if mutatedNode != nil {
				itemFilePath := node.GetAnnotations()["internal.config.kubernetes.io/path"]
				if itemFilePath == "" {
					itemFilePath = node.GetAnnotations()["config.kubernetes.io/path"]
				}

				resultItems = append(resultItems, framework.ResultItem{
					Message: fmt.Sprintf("Annonation: [%s] removed from resource: [%s]", annotationKey, node.GetName()),
					File: framework.File{
						Path: itemFilePath,
					},
					Severity: framework.Info,
				})
			}
		}
	}

	if len(resultItems) > 0 {
		infoResultSlice := []framework.ResultItem{}
		infoResultSlice = append(infoResultSlice, framework.ResultItem{
			Severity: framework.Info,
			Message:  "The following annotations were deleted from the resources",
		})

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, framework.ResultItem{
			Message:  "None of the resources had the provided annotations to delete",
			Severity: framework.Warning,
		})
	}

	return resultItems, nil
}

// getAnnotationKeys gets the keys to delete from resources from the functionConfig
func getAnnotationKeys(fc *yaml.RNode, keys *string) error {
	if len(fc.GetDataMap()) < 1 {
		return errors.New("expecting 1 or more annotation keys to delete as part of the ConfigMap")
	}

	*keys = fc.GetDataMap()[annotationKeysLiteral]

	return nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to process resources: %s", errMsg),
			Severity: framework.Error,
		},
	}
}

type DeleteAnnotationsProcessor struct{}
