// This file will be processed and embedded to pluginator.

package main

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-annotations/generated"
)

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigVersion    = "v1alpha1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "SetAnnotationConfig"
)

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`
}

type setAnnotationFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	plugin             `json:",inline" yaml:",inline"`
}

func (f *setAnnotationFunction) Config(fnConfig interface{}) error {
	configMap, ok := fnConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("function config %#v is not valid", fnConfig)
	}
	rn, err := kyaml.FromMap(configMap)
	if err != nil {
		return fmt.Errorf("failed to construct RNode from %#v: %w", configMap, err)
	}
	switch {
	case f.validGVK(rn, "v1", "ConfigMap"):
		f.plugin.Annotations = rn.GetDataMap()
	case f.validGVK(rn, fnConfigAPIVersion, fnConfigKind):
		// input config is a CRD
		y, err := rn.String()
		if err != nil {
			return fmt.Errorf("cannot get YAML from RNode: %w", err)
		}
		err = yaml.Unmarshal([]byte(y), &f.plugin)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config %#v: %w", y, err)
		}
	default:
		return fmt.Errorf("function config must be a ConfigMap or %s", fnConfigKind)
	}

	if len(f.plugin.Annotations) == 0 {
		return fmt.Errorf("input annotation list cannot be empty")
	}
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}
	// append default field specs
	f.plugin.FieldSpecs = append(f.plugin.FieldSpecs, tc.FieldSpecs...)
	return nil
}

func (f *setAnnotationFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromRNodeSlice(items)
	if err != nil {
		return nil, err
	}
	err = f.plugin.Transform(resMap)
	if err != nil {
		return nil, fmt.Errorf("failed to run transformer: %w", err)
	}
	return resMap.ToRNodeSlice()
}

func (f *setAnnotationFunction) validGVK(rn *kyaml.RNode, apiVersion, kind string) bool {
	meta, err := rn.GetMeta()
	if err != nil {
		return false
	}
	if meta.APIVersion != apiVersion || meta.Kind != kind {
		return false
	}
	return true
}

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["commonannotations"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	return resmap.NewFactory(resourceFactory, nil)
}

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		err := run(resourceList)
		if err != nil {
			resourceList.Result = &framework.Result{
				Name: "set-annotations",
				Items: []framework.Item{
					{
						Message:  err.Error(),
						Severity: framework.Error,
					},
				},
			}
			return resourceList.Result
		}
		return nil
	})

	cmd.Short = generated.SetAnnotationsShort
	cmd.Long = generated.SetAnnotationsLong
	cmd.Example = generated.SetAnnotationsExamples
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(resourceList *framework.ResourceList) error {
	var fn setAnnotationFunction
	err := fn.Config(resourceList.FunctionConfig)
	if err != nil {
		return fmt.Errorf("failed to configure function: %w", err)
	}

	resourceList.Items, err = fn.Run(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to run function: %w", err)
	}
	return nil
}
