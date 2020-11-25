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
	defaultConfigString := `
- path: metadata/namespace
  create: true`
	var defaultConfig []types.FieldSpec
	err := yaml.Unmarshal([]byte(defaultConfigString), &defaultConfig)
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

Example:

To add a namespace 'color' to all resources:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  namespace: color

You can use key "fieldSpecs" to specify the resource selector you
want to use. By default, the function will use this field spec:

- path: metadata/namespace
  create: true

This means a "metadata/namespace" field will be added to all resources
with namespaceable kinds. This information is collected from Kubernetes
openAPI schema. "create: true" means a field 'metadata/namespace' will
be created if it doesn't exist. To support your own CRDs you will beed
to add more items to fieldSpecs list.

"RoleBinding" and "ClusterRoleBinding" will be handled specially because
the "subject" fields in these 2 kinds of resources have references to
namespace names.

For more information about fieldSpecs, please see 
https://kubectl.docs.kubernetes.io/guides/extending_kustomize/builtins/#arguments-4

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
`
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
