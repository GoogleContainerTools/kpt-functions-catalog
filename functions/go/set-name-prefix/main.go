package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-name-prefix/nameref"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`
}

type setNamePrefixSpec struct {
	Prefix     string            `json:"prefix,omitempty" yaml:"name,omitempty"`
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

func setPrefix(spec setNamePrefixSpec,
	resMap resmap.ResMap,
	tc transformerConfig,
	pluginHelpers *resmap.PluginHelpers,
	plugin *plugin) error {
	if spec.Prefix == "" {
		return fmt.Errorf("prefix cannot be empty")
	}

	err := plugin.Config(pluginHelpers, []byte{})
	if err != nil {
		return fmt.Errorf("failed to config plugin: %w", err)
	}
	// append default field specs
	plugin.FieldSpecs = append(spec.FieldSpecs, tc.FieldSpecs...)
	// set name prefix
	plugin.Prefix = spec.Prefix

	err = plugin.Transform(resMap)
	if err != nil {
		return fmt.Errorf("failed to run transformer: %w", err)
	}
	return nil
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
			return fmt.Errorf("failed to convert items to resource map: %w", err)
		}
		prefix, err := getPrefix(resourceList.FunctionConfig)
		if err != nil {
			return fmt.Errorf("failed to get data.specs field from function config: %w", err)
		}
		err = setPrefix(prefix, resMap, tc, pluginHelpers, plugin)
		if err != nil {
			return fmt.Errorf("failed to add name prefix %s: %w",
				prefix.Prefix, err)
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

	cmd.Long = usage()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func usage() string {
	return `Add a name prefix to all resources.

Configured using a ConfigMap with the following key in 'data':

prefix: Name prefix to add to resources.

Example:

To add a name prefix 'dev-' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  prefix: dev-

Each function call can only add one name prefix. This function
will keep idempotent across different calls. If a name of a
resource has already had the same prefix then it will not add
the prefix again.

You can use key 'fieldSpecs' to specify the resource selector you
want to use. By default, the function will not only add or update the
labels in 'metadata/name' but also a bunch of different places where
have references to the names. These field specs are defined in
https://github.com/kubernetes-sigs/kustomize/blob/master/api/konfig/builtinpluginconsts/namereference.go

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

To add a name 'dev-' to path 'data/selector/name' in MyOwnResource:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  prefix: dev-
  fieldSpecs:
  - path: data/selector/name
    kind: MyOwnKind
`
}

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["nameprefix"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}

func newPluginHelpers(resmapFactory *resmap.Factory) *resmap.PluginHelpers {
	return resmap.NewPluginHelpers(nil, nil, resmapFactory)
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := provider.NewDefaultDepProvider().GetResourceFactory()
	return resmap.NewFactory(resourceFactory, nil)
}

func getPrefix(fc interface{}) (setNamePrefixSpec, error) {
	var fcd setNamePrefixSpec
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
	err = yaml.Unmarshal([]byte(b), &fcd)
	if err != nil {
		return fcd, err
	}
	return fcd, nil
}
