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
	fnConfigKind       = "SetNamespaceConfig"
)

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type setNamespaceConfig struct {
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
				Name: "set-namespace",
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
	defaultConfig, err := getDefaultConfig()
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
		return fmt.Errorf("failed to config plugin: %w", err)
	}
	if plugin.Namespace == "" {
		return fmt.Errorf("namespace in the input config cannot be empty")
	}
	if len(plugin.FieldSpecs) == 0 {
		plugin.FieldSpecs = defaultConfig.FieldSpecs
	}
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
	return `Update or add namespace.

To add a namespace 'color' to all resources, configure the function
with a simple ConfigMap like this:

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  namespace: color

There is another advanced way to configure the function by a custom
resource which can provide more flexibility than using ConfigMap.

Example:

apiVersion: kpt.dev/v1
kind: SetNamespaceConfig
metadata:
  name: my-config
namespace: color

The values for 'apiVersion' and 'kind' must match the values in the
example above.

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

apiVersion: kpt.dev/v1
kind: SetNamespaceConfig
metadata:
  name: my-config
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
func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["namespace"]
	var defaultConfig transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &defaultConfig)
	return defaultConfig, err
}

//nolint
func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	return resmap.NewFactory(resourceFactory, nil)
}

//nolint
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
		plugin.Namespace = data["namespace"]
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
	config := setNamespaceConfig{}
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
