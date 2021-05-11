package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/upsert-resource/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/upsert-resource/upsertresource"
	"github.com/ghodss/yaml"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		ur, err := getUpsertResource(resourceList.FunctionConfig)
		if err != nil {
			return fmt.Errorf("failed to parse function config: %w", err)
		}
		resourceList.Items, err = ur.Filter(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to upsert resource: %w", err)
		}
		return nil
	})

	cmd.Short = generated.UpsertResourceShort
	cmd.Long = generated.UpsertResourceLong
	cmd.Example = generated.UpsertResourceExamples

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getUpsertResource gets the UpsertResource instance with input Resource to upsert
func getUpsertResource(fnConfig interface{}) (*upsertresource.UpsertResource, error) {
	config, err := yaml.Marshal(fnConfig)
	if err != nil {
		return nil, err
	}
	configNode, err := kyaml.Parse(string(config))
	if err != nil {
		return nil, err
	}
	// format the resource as is it parsed from interface{} type
	_, err = filters.FormatFilter{UseSchema: true}.Filter([]*kyaml.RNode{configNode})
	if err != nil {
		return nil, err
	}
	ur := &upsertresource.UpsertResource{
		Resource: configNode,
	}
	return ur, nil
}
