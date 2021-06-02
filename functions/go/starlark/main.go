package main

import (
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/starlark/generated"
)

func main() {
	sf := &StarlarkRun{}
	resourceList := &framework.ResourceList{
		FunctionConfig: sf,
	}

	cmd := framework.Command(resourceList, func() error {
		err := func() error {
			if ve := sf.Validate(); ve != nil {
				return ve
			}
			if te := sf.Transform(resourceList); te != nil {
				return te
			}
			return nil
		}()
		if err != nil {
			resourceList.Result = &framework.Result{
				Name: "starlark",
				Items: []framework.Item{
					{
						Message:  err.Error(),
						Severity: framework.Error,
					},
				},
			}
			resourceList.FunctionConfig = nil
			return resourceList.Result
		}
		return nil
	})
	cmd.Short = generated.StarlarkShort
	cmd.Long = generated.StarlarkLong
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
