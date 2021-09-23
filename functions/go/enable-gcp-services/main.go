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
	psf := ProjectServiceSetFunction{}
	cmd := command.Build(&psf, command.StandaloneEnabled, false)

	cmd.Short = generated.EnableGcpServicesShort
	cmd.Long = generated.EnableGcpServicesLong
	cmd.Example = generated.EnableGcpServicesExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ProjectServiceSetFunction struct{}

func (psf *ProjectServiceSetFunction) Process(resourceList *framework.ResourceList) error {
	var pslr gcpservices.ProjectServiceSetRunner

	resourceList.Result = &framework.Result{
		Name: "enable-gcp-services",
	}
	var err error
	resourceList.Items, err = pslr.Filter(resourceList.Items)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = pslr.GetResults()
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
