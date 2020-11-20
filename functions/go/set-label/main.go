package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/yaml"
)

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`
}

type functionConfigData struct {
	LabelName  string `json:"label_name,omitempty" yaml:"label_name,omitempty"`
	LabelValue string `json:"label_value,omitempty" yaml:"label_value,omitempty"`
}

//nolint
func main() {
	var plugin *plugin = &KustomizePlugin
	tc, err := getDefaultConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resmapFactory := newResMapFactory()

	pluginHelpers := newPluginHelpers(resmapFactory)

	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
		if err != nil {
			return errors.Wrap(err, "failed to convert items to resource map")
		}
		data, err := getData(resourceList.FunctionConfig)
		if err != nil {
			return errors.Wrap(err, "failed to get data field from function config")
		}
		if data.LabelName == "" || data.LabelValue == "" {
			return fmt.Errorf("label_name and label_value cannot be empty")
		}
		dataBytes, err := getDataBytes(resourceList.FunctionConfig)
		if err != nil {
			return errors.Wrap(err, "failed to get data bytes")
		}

		err = plugin.Config(pluginHelpers, dataBytes)
		if err != nil {
			return errors.Wrap(err, "failed to config plugin")
		}
		// append default field specs
		plugin.FieldSpecs = append(plugin.FieldSpecs, tc.FieldSpecs...)
		// set label key and value
		plugin.Labels = make(map[string]string)
		plugin.Labels[data.LabelName] = data.LabelValue

		err = plugin.Transform(resMap)
		if err != nil {
			return errors.Wrap(err, "failed to run transformer")
		}

		resourceList.Items, err = resMap.ToRNodeSlice()
		if err != nil {
			return errors.Wrap(err, "failed to convert resource map to items")
		}
		return nil
	})

	cmd.Long = usage()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func usage() string {
	return `Add a label to all resources.

Configured using a ConfigMap with the following keys:

label_name: Label name to add to resources.
label_value: Label value to add to resources.

Example:

To add a label 'color: orange' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  label_name: color
  label_value: orange

You can use key 'fieldSpecs' to specify the resource selector you
want to use. By default, the function will not only add or update the
labels in 'metadata/labels' but also a bunch of different places where
have references to the labels. These field specs are defined in
https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/commonlabels.go#L6

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

Example:

To add a label 'color: orange' to path 'data/selector' in MyOwnKind resource:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  label_name: color
  label_value: orange
  fieldSpecs:
  - path: data/selector
	kind: MyOwnKind
	create: true
`
}

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["commonlabels"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}

func newPluginHelpers(resmapFactory *resmap.Factory) *resmap.PluginHelpers {
	return resmap.NewPluginHelpers(nil, nil, resmapFactory)
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	return resmap.NewFactory(resourceFactory, nil)
}

func getData(fc interface{}) (functionConfigData, error) {
	var fcd functionConfigData
	b, err := getDataBytes(fc)
	if err != nil {
		return fcd, err
	}

	err = yaml.Unmarshal(b, &fcd)
	if err != nil {
		return fcd, err
	}
	return fcd, nil
}

func getDataBytes(fc interface{}) ([]byte, error) {
	f, ok := fc.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("function config %#v is not valid", fc)
	}
	b, err := yaml.Marshal(f["data"])
	if err != nil {
		return nil, err
	}
	return b, nil
}
