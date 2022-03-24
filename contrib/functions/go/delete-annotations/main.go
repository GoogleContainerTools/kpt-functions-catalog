package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

const (
	annotationKeysLiteral = "annotationKeys"
	annotationDelimeter   = ","
)

//nolint
func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(deleteAnnotations)); err != nil {
		os.Exit(1)
	}
}

func deleteAnnotations(rl *fn.ResourceList) error {

	var annotationKeys []string
	annotationKeys, err := getAnnotationKeys(rl.FunctionConfig)

	if err != nil {
		rl.Results = append(rl.Results, getErrorItem(err.Error()))
		return err
	}

	items, err := processResources(rl.Items, annotationKeys)
	if err != nil {
		rl.Results = append(rl.Results, getErrorItem(err.Error()))
		return err
	}

	for _, resultItem := range items {
		rl.Results = append(rl.Results, resultItem)
	}

	return nil
}

func processResources(objects []*fn.KubeObject, annotationKeys []string) ([]*fn.Result, error) {
	var resultItems []*fn.Result
	for _, o := range objects {
		if o.GetKind() == "" || o.GetName() == "" || o.GetAPIVersion() == "" {
			continue
		}

		for _, annotationKey := range annotationKeys {
			removed, err := o.Remove("metadata", "annotations", annotationKey)
			if err != nil {
				return nil, err
			}

			if removed {
				itemFilePath := o.GetAnnotations()["internal.config.kubernetes.io/path"]
				if itemFilePath == "" {
					itemFilePath = o.GetAnnotations()["config.kubernetes.io/path"]
				}

				resultItems = append(resultItems, &fn.Result{
					Message: fmt.Sprintf("Annonation: [%s] removed from resource: [%s]", annotationKey, o.GetName()),
					File: &fn.File{
						Path: itemFilePath,
					},
					Severity: fn.Info,
				})
			}
		}
	}

	if len(resultItems) > 0 {
		infoResultSlice := []*fn.Result{}
		infoResultSlice = append(infoResultSlice, &fn.Result{
			Severity: fn.Info,
			Message:  "The following annotations were deleted from the resources",
		})

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, &fn.Result{
			Message:  "None of the resources had the provided annotations to delete",
			Severity: fn.Warning,
		})
	}

	return resultItems, nil
}

// getAnnotationKeys gets the keys to delete from resources from the functionConfig
func getAnnotationKeys(fc *fn.KubeObject) ([]string, error) {
	annotationKeysString := fc.GetStringOrDie("data", annotationKeysLiteral)

	if annotationKeysString == "" {
		return nil, fmt.Errorf("%s was not provided as part of the config or paramters to the function", annotationKeysLiteral)
	}

	annotationKeys := strings.Split(strings.TrimSpace(annotationKeysString), annotationDelimeter)
	return annotationKeys, nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) *fn.Result {
	return &fn.Result{
		Message:  fmt.Sprintf("failed to process resources: %s", errMsg),
		Severity: fn.Error,
	}
}
