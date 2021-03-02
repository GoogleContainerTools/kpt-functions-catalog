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
	fnConfigKind       = "SetLabelConfig"
)

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`
}

type setLabelConfig struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	plugin             `json:",inline" yaml:",inline"`
}

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		err := run(resourceList)
		if err != nil {
			resourceList.Result = &framework.Result{
				Name: "set-label",
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
	var plugin *plugin = &KustomizePlugin
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}
	resmapFactory := newResMapFactory()

	resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to convert items to resource map: %w", err)
	}
	err = configTransformer(resourceList.FunctionConfig, plugin)
	if err != nil {
		return fmt.Errorf("failed to configure transformer: %w", err)
	}
	if len(plugin.Labels) == 0 {
		return fmt.Errorf("input label list cannot be empty")
	}
	// append default field specs
	plugin.FieldSpecs = append(plugin.FieldSpecs, tc.FieldSpecs...)
	err = plugin.Transform(resMap)
	if err != nil {
		return fmt.Errorf("failed to run transformer: %w", err)
	}

	resourceList.Items, err = resMap.ToRNodeSlice()
	if err != nil {
		return fmt.Errorf("failed to convert resource map to items: %w", err)
	}
	return nil
}

func usage() string {
	return `Add a list of labels to all resources.

Configured using a ConfigMap with key-value pairs in 'data' field in
'ConfigMap' resource. Example:

To add a label 'color: orange' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange

To add 2 labels 'color: orange' and 'fruit: apple' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  color: orange
  fruit: apple

You can use key 'fieldSpecs' to specify the resource selector you
want to use. By default, the function will not only add or update the
labels in 'metadata/labels' but also a bunch of different places where
have references to the labels. These field specs are defined in
https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonlabels.go#L6

You need to use a custom resource to specify additional information.

Example:

To add a label 'color: orange' to path 'data/selector' in MyOwnKind resource:

apiVersion: kpt.dev/v1
kind: SetLabelConfig
metadata:
  name: my-config
labels:
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

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["commonlabels"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	return resmap.NewFactory(resourceFactory, nil)
}

func configTransformer(fc interface{}, plugin *plugin) error {
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fmt.Errorf("failed to construct RNode from %#v: %w", f, err)
	}
	ok, err = isConfigMap(rn)
	if err != nil {
		return err
	}
	if ok {
		// input config is a ConfigMap
		data := rn.GetDataMap()
		plugin.Labels = make(map[string]string)
		for k, v := range data {
			plugin.Labels[k] = v
		}
		return nil
	}
	ok, err = isCrdConfig(rn)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("function config must be a ConfigMap or %s", fnConfigKind)
	}
	// input config is a CRD
	y, err := rn.String()
	if err != nil {
		return fmt.Errorf("cannot get YAML from RNode: %w", err)
	}
	config := setLabelConfig{}
	err = yaml.Unmarshal([]byte(y), &config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config %#v: %w", y, err)
	}
	*plugin = config.plugin
	return nil
}

func isConfigMap(rn *kyaml.RNode) (bool, error) {
	meta, err := rn.GetMeta()
	if err != nil {
		return false, fmt.Errorf("failed to get metadata: %w", err)
	}
	if meta.APIVersion != "v1" || meta.Kind != "ConfigMap" {
		return false, nil
	}
	return true, nil
}

func isCrdConfig(rn *kyaml.RNode) (bool, error) {
	meta, err := rn.GetMeta()
	if err != nil {
		return false, fmt.Errorf("failed to get metadata: %w", err)
	}
	if meta.APIVersion != fnConfigAPIVersion || meta.Kind != fnConfigKind {
		return false, nil
	}
	return true, nil
}
