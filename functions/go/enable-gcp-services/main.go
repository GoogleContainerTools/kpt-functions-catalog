package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/project-services/gcpservices"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/project-services/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

//nolint
func main() {
	psf := ProjectServiceListFunction{}
	cmd := command.Build(&psf, command.StandaloneEnabled, false)

	cmd.Short = generated.ProjectServiceListShort
	cmd.Long = generated.ProjectServiceListLong
	cmd.Example = generated.ProjectServiceListExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ProjectServiceListFunction struct{}

func (psf *ProjectServiceListFunction) Process(resourceList *framework.ResourceList) error {
	var psl gcpservices.ProjectServiceList
	err := framework.LoadFunctionConfig(resourceList.FunctionConfig, &psl)
	if err != nil {
		return fmt.Errorf("failed to load the `functionConfig`: %w", err)
	}

	resourceList.Result = &framework.Result{
		Name: "enable-gcp-services",
	}
	resourceList.Items, err = psl.Filter(resourceList.Items)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}

	results := psl.GetResults()
	for _, r := range results {
		resourceList.Result.Items = append(resourceList.Result.Items,
			framework.ResultItem{
				Message:     r.Action,
				Severity:    framework.Info,
				ResourceRef: r.ResourceRef,
				File:        framework.File{Path: r.FilePath},
			})
	}
	return nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to add services: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
