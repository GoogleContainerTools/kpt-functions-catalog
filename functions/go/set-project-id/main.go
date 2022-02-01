package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const projectIDKey = "projectID"


type Processor struct {}

func getProjectID(resourceList *framework.ResourceList) (string, error) {
	data := resourceList.FunctionConfig.GetDataMap()
	if data == nil {
		return "", fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	projectID, ok := data[projectIDKey]
	if !ok {
		return "", fmt.Errorf("missing `.data.%s` field in `ConfigMap` FunctionConfig", projectIDKey)
	}
	return projectID, nil
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	return resmap.NewFactory(resourceFactory)
}


func (p *Processor) Process(resourceList *framework.ResourceList) error {
	err := func() error{
		// FunctionConfig is ConfigMap kind. No need for Validator, Defaultor struct
		projectID, err := getProjectID(resourceList)
		if err != nil {
			return err
		}
		var trans ProjectIDTransformer
		if err := yaml.Unmarshal([]byte(projectIDFieldSpecs), &trans); err != nil {
			return err
		}
		resmapFactory := newResMapFactory()
		resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to convert items to resource map: %w", err)
		}
		if err := trans.Transform(resMap, projectID); err != nil {
			return err
		}
		resourceList.Items = resMap.ToRNodeSlice()
		return nil
	}()
	if err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	return nil
}

func main() {
	cmd := command.Build(&Processor{}, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
