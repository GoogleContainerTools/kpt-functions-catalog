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
)

const (
	fnConfigGroup      = "kpt.dev"
	fnConfigVersion    = "v1"
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
				Name: "set-annotation",
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

	cmd.Long = usage()
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

func usage() string {
	return `Add a list of annotations to all resources.

Configured using a ConfigMap with key-value pairs in 'data' field in
'ConfigMap' resource. Example:

To add a annotation 'color: orange' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange

To add 2 annotations 'color: orange' and 'fruit: apple' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange
  fruit: apple

You can use key 'fieldSpecs' to specify the resource selector you
want to use. By default, the function will not only add or update the
annotations in 'metadata/annotations' but also a bunch of different places where
have references to the annotations. These field specs are defined in
https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonannotations.go#L6

You need to use a custom resource to specify additional information.

Example:

To add an annotation 'color: orange' to path 'data/selector' in MyOwnKind resource:

apiVersion: kpt.dev/v1
kind: SetAnnotationConfig
metadata:
  name: my-config
annotations:
  color: orange
fieldSpecs:
- path: data/selector
  kind: MyOwnKind
  create: true

To support your own CRDs you will need to add more items to fieldSpecs list.
Your own specs will be used with the default ones.

Field spec has following fields:

- group: Select the resources by API version group. Will select all groups
	if omitted.
- version: Select the resources by API version. Will select all versions
	if omitted.
- kind: Select the resources by resource kind. Will select all kinds
	if omitted.
- path: Specify the path to the field that the value will be updated. This field
	is required.
- create: If it's set to true, the field specified will be created if it doesn't
	exist. Otherwise the function will only update the existing field.

For more information about fieldSpecs, please see 
https://kubectl.docs.kubernetes.io/guides/extending_kustomize/builtins/#arguments-3
`
}
