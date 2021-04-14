package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/yaml"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-name-prefix/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-name-prefix/nameref"
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

	ens := &EnsureNameSegment{}
	resourceList := &framework.ResourceList{
		FunctionConfig: ens,
	}

	cmd := framework.Command(resourceList, func() error {
		ens.FieldSpecs = append(tc.FieldSpecs, ens.FieldSpecs...)

		resourceFactory := provider.NewDefaultDepProvider().GetResourceFactory()
		resmapFactory := resmap.NewFactory(resourceFactory, nil)

		resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
		if err != nil {
			return fmt.Errorf("failed to convert items to resource map: %w", err)
		}

		ens.Defaults()
		if err = ens.Validate(); err != nil {
			return fmt.Errorf("failed validation: %w", err)
		}
		if err = ens.Transform(resMap); err != nil {
			return fmt.Errorf("failed to transform name segment: %w", err)
		}
		// update name back reference
		err = nameref.FixNameBackReference(resMap)
		if err != nil {
			return fmt.Errorf("failed to fix name back reference: %w", err)
		}

		// remove kustomize build annotations
		resMap.RemoveBuildAnnotations()
		resourceList.Items, err = resMap.ToRNodeSlice()
		if err != nil {
			return fmt.Errorf("failed to convert resource map to items: %w", err)
		}
		return nil
	})

	cmd.Short = generated.EnsureNameSegmentShort
	cmd.Long = generated.EnsureNameSegmentLong
	cmd.Example = generated.EnsureNameSegmentExamples
	return cmd.Execute()
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
