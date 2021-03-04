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
	fnConfigVersion    = "v1beta1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "SetNamespaceConfig"
)

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type setNamespaceFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	plugin             `json:",inline" yaml:",inline"`
}

func (f *setNamespaceFunction) Config(fnConfig interface{}) error {
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
		f.plugin.Namespace = rn.GetDataMap()["namespace"]
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

	if f.plugin.Namespace == "" {
		return fmt.Errorf("input namespace cannot be empty")
	}
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}
	// set default field specs
	f.plugin.FieldSpecs = append(f.plugin.FieldSpecs, tc.FieldSpecs...)
	return nil
}

func (f *setNamespaceFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
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

func (f *setNamespaceFunction) validGVK(rn *kyaml.RNode, apiVersion, kind string) bool {
	meta, err := rn.GetMeta()
	if err != nil {
		return false
	}
	if meta.APIVersion != apiVersion || meta.Kind != kind {
		return false
	}
	return true
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
	var fn setNamespaceFunction
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

apiVersion: kpt.dev/v1beta1
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

To support your own CRDs you will need to add more items to fieldSpecs list.

In addition to simple updating the 'metadata/namespace' fields for resources, it's
also possible that references to namespace names exist in some resources. In this
case, these references will also need to be updated at the same time. In Kubernetes
native resources, the references to namespace in 'subject' fields in 'RoleBinding'
and 'ClusterRoleBinding' will be handled implicitly by this function. If you have
references to namespaces in you CRDs, use the field spec described above to update
them.

Fieldspec has following fields:

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

To add a namespace 'color' to 'spec/selector/namespace' in 'MyCRD' resource:

apiVersion: kpt.dev/v1beta1
kind: SetNamespaceConfig
metadata:
  name: my-config
namespace: color
fieldSpecs:
- path: spec/selector/namespace
  kind: MyCRD
  version: v1
  group: example.com
  create: true

For more information about fieldSpecs, please see 
https://kubectl.docs.kubernetes.io/guides/extending_kustomize/builtins/#arguments-4
`
}
