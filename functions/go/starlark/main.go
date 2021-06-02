package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/starlark/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type StarlarkProcessor struct{}

func (gkp *StarlarkProcessor) Process(resourceList *framework.ResourceList) error {
	err := func() error {
		s, err := resourceList.FunctionConfig.String()
		if err != nil {
			return err
		}
		sf := StarlarkRun{}
		if err = yaml.Unmarshal([]byte(s), &sf); err != nil {
			return err
		}
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
			Items: []framework.ResultItem{
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
}

func main() {
	sp := StarlarkProcessor{}
	cmd := command.Build(&sp, command.StandaloneEnabled, false)
	cmd.Short = generated.StarlarkShort
	cmd.Long = generated.StarlarkLong
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
