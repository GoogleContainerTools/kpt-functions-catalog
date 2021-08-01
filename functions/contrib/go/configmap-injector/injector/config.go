package injector

import (
	"encoding/json"
	"fmt"
	"strings"

	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	fnName             = "configmap-injector"
	fnConfigGroup      = "fn.kumorilabs.io"
	fnConfigVersion    = "v1alpha1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "ConfigMapInjector"
)

type ConfigMapInjector struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec               Spec `yaml:"spec,omitempty"`
}

type Spec struct {
	// Target is the ConfigMap KRM to be updated with the values and template
	Target *ConfigMapRef `yaml:"target,omitempty"`
	// Values specifies a ConfigMap KRM where setters can be applied. Each value will be applyed to the ConfigMap specified in template
	Values *ConfigMapRef `yaml:"values,omitempty"`
	// Template specifies a ConfigMap KRM which contains the config template
	Template *ConfigMapRef `yaml:"template,omitempty"`
}

type ConfigMapRef struct {
	kyaml.NameMeta
}

func (f *ConfigMapInjector) Validate(fnConfig interface{}) error {

	if f.Spec.Template == nil {
		return fmt.Errorf("input Template cannot be empty")
	}

	if f.Spec.Target == nil {
		return fmt.Errorf("input Target cannot be empty")
	}

	if f.Spec.Values == nil {
		return fmt.Errorf("input Values cannot be empty")
	}

	return nil
}

// Filter implements Set as a yaml.Filter
func (f *ConfigMapInjector) Filter(items []*kyaml.RNode) ([]*kyaml.RNode, error) {

	templateRes := &kyaml.RNode{}
	valuesRes := &kyaml.RNode{}
	targetRes := &kyaml.RNode{}

	for _, res := range items {
		if res.GetKind() != "ConfigMap" {
			continue
		}

		name := res.GetName()

		if name == f.Spec.Target.Name {
			targetRes = res
		}

		if name == f.Spec.Template.Name {
			templateRes = res
		}

		if name == f.Spec.Values.Name {
			valuesRes = res
		}

	}

	if targetRes == nil {
		return nil, fmt.Errorf("target resource with name [%s] not present", f.Spec.Target.Name)
	}

	if templateRes == nil {
		return nil, fmt.Errorf("template resource with name [%s] not present", f.Spec.Template.Name)
	}

	if valuesRes == nil {
		return nil, fmt.Errorf("values resource with name [%s] not present", f.Spec.Values.Name)
	}

	targetMap := targetRes.GetDataMap()
	templateMap := templateRes.GetDataMap()
	valuesMap := valuesRes.GetDataMap()

	for k, v := range templateMap {
		targetMap[k] = replaceValues(v, valuesMap)
	}

	targetRes.SetDataMap(targetMap)

	return items, nil
}

func replaceValues(data string, values map[string]string) string {
	for k, v := range values {
		sub := "${" + k + "}"
		data = strings.ReplaceAll(data, sub, v)
	}
	return data
}

// Decode decodes the input yaml node into ConfigMapInjector struct
func Decode(rn *kyaml.RNode, fcd *ConfigMapInjector) {
	j, _ := rn.MarshalJSON()
	json.Unmarshal(j, fcd)
}
