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

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`
}

type setLabelSpecs struct {
	Labels []setLabelSpec `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type setLabelSpec struct {
	LabelName  string            `json:"name,omitempty" yaml:"name,omitempty"`
	LabelValue string            `json:"value,omitempty" yaml:"value,omitempty"`
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

func setLabel(spec setLabelSpec,
	resMap resmap.ResMap,
	tc transformerConfig,
	pluginHelpers *resmap.PluginHelpers,
	plugin *plugin) error {
	if spec.LabelName == "" || spec.LabelValue == "" {
		return fmt.Errorf("labels.name and labels.value cannot be empty")
	}

	err := plugin.Config(pluginHelpers, []byte{})
	if err != nil {
		return fmt.Errorf("failed to config plugin: %w", err)
	}
	// append default field specs
	plugin.FieldSpecs = append(spec.FieldSpecs, tc.FieldSpecs...)
	// set label key and value
	plugin.Labels = make(map[string]string)
	plugin.Labels[spec.LabelName] = spec.LabelValue

	err = plugin.Transform(resMap)
	if err != nil {
		return fmt.Errorf("failed to run transformer: %w", err)
	}
	return nil
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
	pluginHelpers := newPluginHelpers(resmapFactory)

	resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to convert items to resource map: %w", err)
	}
	labels, err := getLabels(resourceList.FunctionConfig)
	if err != nil {
		return fmt.Errorf("failed to get data.specs field from function config: %w", err)
	}
	if len(labels.Labels) == 0 {
		return fmt.Errorf("input label list cannot be empty")
	}
	for _, l := range labels.Labels {
		err := setLabel(l, resMap, tc, pluginHelpers, plugin)
		if err != nil {
			return fmt.Errorf("failed to add label [%s: %s]: %w",
				l.LabelName, l.LabelValue, err)
		}
	}

	resourceList.Items, err = resMap.ToRNodeSlice()
	if err != nil {
		return fmt.Errorf("failed to convert resource map to items: %w", err)
	}
	return nil
}

func usage() string {
	return `Add a list of labels to all resources.

Configured using a ConfigMap with the following keys:

labels.name: Label name to add to resources.
labels.value: Label value to add to resources.

These keys are in a list in path 'data.labels'.

Example:

To add a label 'color: orange' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  labels:
  - name: color
    value: orange

  To add 2 labels 'color: orange' and 'fruit: apple' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  labels:
  - name: color
    value: orange
  - name: fruit
    value: apple

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
  labels:
    - name: color
      value: orange
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

func getLabels(fc interface{}) (setLabelSpecs, error) {
	var fcd setLabelSpecs
	f, ok := fc.(map[string]interface{})
	if !ok {
		return fcd, fmt.Errorf("function config %#v is not valid", fc)
	}
	rn, err := kyaml.FromMap(f)
	if err != nil {
		return fcd, fmt.Errorf("failed to parse from function config: %w", err)
	}
	specsNode, err := rn.Pipe(kyaml.Lookup("data"))
	if err != nil {
		return fcd, err
	}

	b, err := specsNode.String()
	if err != nil {
		return fcd, err
	}
	// check does data contains key-value pairs
	var keyValueMap map[string]string
	err = yaml.Unmarshal([]byte(b), &keyValueMap)
	if err == nil {
		// we got a simple key-value pair
		for k, v := range keyValueMap {
			fcd.Labels = append(fcd.Labels,
				setLabelSpec{LabelName: k, LabelValue: v})
		}
		return fcd, nil
	}

	err = yaml.Unmarshal([]byte(b), &fcd)
	if err != nil {
		return fcd, err
	}
	return fcd, nil
}
