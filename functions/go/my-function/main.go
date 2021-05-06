package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/my-function/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/my-function/myfunction"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		s, err := getParams(resourceList.FunctionConfig)
		if err != nil {
			return fmt.Errorf("failed to parse function config: %w", err)
		}
		_, err = s.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to run function: %w", err)
		}
		return nil
	})

	cmd.Short = generated.MyFunctionShort
	cmd.Long = generated.MyFunctionLong
	cmd.Example = generated.MyFunctionExamples

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getParams retrieve the parameters from input config
func getParams(fc interface{}) (myfunction.FunctionConfig, error) {
	var fcd myfunction.FunctionConfig
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fcd, fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fcd, fmt.Errorf("failed to parse input from function config: %w", err)
	}
	myfunction.Decode(rn, &fcd)
	return fcd, nil
}
