package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/yaml"
)

//nolint
func main() {
	var plugin *plugin = &KustomizePlugin
	defaultConfig, err := getDefaultConfig()
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
		dataField, err := getDataFromFunctionConfig(resourceList.FunctionConfig)
		if err != nil {
			return errors.Wrap(err, "failed to get data field from function config")
		}
		dataValue, err := yaml.Marshal(dataField)
		if err != nil {
			return errors.Wrap(err, "error when marshal data values")
		}

		err = plugin.Config(pluginHelpers, dataValue)
		if err != nil {
			return errors.Wrap(err, "failed to config plugin")
		}
		if len(plugin.FieldSpecs) == 0 {
			plugin.FieldSpecs = defaultConfig
		}
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
	return `Update or add namespace.

Configured using a ConfigMap with the following keys:

namespace: Name of the namespace that will be set.
fieldSpecs: A list of specification to select the resources and fields that 
the namespace will be applied to.

Example:

To add a namespace 'color' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  namespace: color

You can use key 'fieldSpecs' to specify the resource selector you
want to use. By default, the function will use this field spec:

- path: metadata/namespace
  create: true

This means a 'metadata/namespace' field will be added to all resources
with namespaceable kinds. Whether a resource is namespaceable is determined
by the Kubernetes API schema. If the API path for that kind contains
'namespaces/{namespace}' then the resource is considered namespaceable. Otherwise
it's not. Currently this function is using API version 1.19.1. 

For more information about API schema used in this function, please take a look at
https://github.com/kubernetes-sigs/kustomize/tree/master/kyaml/openapi

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

Example:

To add a namespace 'color' to Deployment resource only:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  namespace: color
  fieldSpecs:
    - path: metadata/namespace
      kind: Deployment
      create: true

For more information about fieldSpecs, please see 
https://kubectl.docs.kubernetes.io/guides/extending_kustomize/builtins/#arguments-4

To support your own CRDs you will need to add more items to fieldSpecs list.

In addition to simple updating the 'metadata/namespace' fields for resources, it's
also possible that references to namespace names exist in some resources. In this
case, these references will also need to be updated at the same time. In Kubernetes
native resources, the references to namespace in 'subject' fields in 'RoleBinding'
and 'ClusterRoleBinding' will be handled implicitly by this function. If you have
references to namespaces in you CRDs, use the field spec described above to update
them.
`
}

//nolint
func getDefaultConfig() ([]types.FieldSpec, error) {
	defaultConfigString := `
- path: metadata/namespace
  create: true`
	var defaultConfig []types.FieldSpec
	err := yaml.Unmarshal([]byte(defaultConfigString), &defaultConfig)
	return defaultConfig, err
}

//nolint
func newPluginHelpers(resmapFactory *resmap.Factory) *resmap.PluginHelpers {
	return resmap.NewPluginHelpers(nil, nil, resmapFactory)
}

//nolint
func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	return resmap.NewFactory(resourceFactory, nil)
}

//nolint
func getDataFromFunctionConfig(fc interface{}) (interface{}, error) {
	f, ok := fc.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("function config %#v is not valid", fc)
	}
	return f["data"], nil
}
