package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/ensure-name-substring/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/ensure-name-substring/nameref"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/yaml"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}

	ensp := EnsureNameSubstringProcessor{
		tc: &tc,
	}
	cmd := command.Build(&ensp, command.StandaloneEnabled, false)

	cmd.Short = generated.EnsureNameSubstringShort
	cmd.Long = generated.EnsureNameSubstringLong
	return cmd.Execute()
}

type EnsureNameSubstringProcessor struct {
	tc *transformerConfig
}

func (ensp *EnsureNameSubstringProcessor) Process(resourceList *framework.ResourceList) error {
	var ens EnsureNameSubstring
	if err := framework.LoadFunctionConfig(resourceList.FunctionConfig, &ens); err != nil {
		return fmt.Errorf("failed to load the `functionConfig`: %w", err)
	}

	if ensp.tc == nil {
		return fmt.Errorf("failed to load the default configuration")
	}

	ens.FieldSpecs = append(ensp.tc.FieldSpecs, ens.FieldSpecs...)

	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	resmapFactory := resmap.NewFactory(resourceFactory)

	resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to convert items to resource map: %w", err)
	}

	ens.Defaults()
	if err = ens.Validate(); err != nil {
		return fmt.Errorf("failed validation: %w", err)
	}
	if err = ens.Transform(resMap); err != nil {
		return fmt.Errorf("failed to transform name substring: %w", err)
	}
	// update name back reference
	err = nameref.FixNameBackReference(resMap)
	if err != nil {
		return fmt.Errorf("failed to fix name back reference: %w", err)
	}

	// remove kustomize build annotations
	resMap.RemoveBuildAnnotations()
	resourceList.Items = resMap.ToRNodeSlice()
	if err != nil {
		return fmt.Errorf("failed to convert resource map to items: %w", err)
	}
	return nil
}

type transformerConfig struct {
	FieldSpecs []types.FieldSpec `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`
}

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["nameprefix"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}
