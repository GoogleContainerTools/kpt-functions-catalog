package injector

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type ConfigmapInjectionProcessor struct{}

func (cip *ConfigmapInjectionProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "configmap-injector",
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
	s, err := getInjector(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}

	err = resourceList.Filter(kio.FilterFunc(s.Filter))
	if err != nil {
		return nil, err
	}

	return resourceList.Result.Items, nil
}

// getInjector retrieves the config from input config
func getInjector(fc *kyaml.RNode) (ConfigMapInjector, error) {
	var fcd ConfigMapInjector
	Decode(fc, &fcd)
	return fcd, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to inject configmap: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
