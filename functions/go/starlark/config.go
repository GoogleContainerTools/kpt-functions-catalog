package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/starlark"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigVersion    = "v1beta1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "StarlarkFunction"
)

type StarlarkFunction struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Source is a required field for providing a starlark script inline.
	Source string `json:"source" yaml:"source"`
	// KeyValues is a convenient way to pass in arbitrary key value pairs.
	Data map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

func (sf *StarlarkFunction) Validate() error {
	if sf.APIVersion != fnConfigAPIVersion {
		return fmt.Errorf("`apiVersion` must be: %v", fnConfigAPIVersion)
	}
	if sf.Kind != fnConfigKind {
		return fmt.Errorf("`kind` must be: %v", fnConfigKind)
	}

	if sf.ObjectMeta.Name == "" {
		return fmt.Errorf("`metadata.name` must be set in starlark function config")
	}

	if sf.Source == "" {
		return fmt.Errorf("`source` must not be empty")
	}
	return nil
}

func (sf *StarlarkFunction) Transform(rl *framework.ResourceList) error {
	err := sf.filterStarlarkFunctionKind(rl)
	if err != nil {
		return err
	}

	fc, err := sf.toRNode()
	if err != nil {
		return err
	}

	starFltr := &starlark.Filter{
		Name:    sf.Name,
		Program: sf.Source,
		FunctionFilter: runtimeutil.FunctionFilter{
			FunctionConfig: fc,
		},
	}
	rl.Items, err = starFltr.Filter(rl.Items)
	return err
}

func (sf *StarlarkFunction) filterStarlarkFunctionKind(rl *framework.ResourceList) error {
	var updated []*yaml.RNode
	for i, item := range rl.Items {
		rm, err := item.GetMeta()
		if err != nil {
			return err
		}
		if rm.Kind == fnConfigKind {
			continue
		}
		updated = append(updated, rl.Items[i])
	}
	rl.Items = updated
	return nil
}

func (sf *StarlarkFunction) toRNode() (*yaml.RNode, error) {
	y, err := yaml.Marshal(sf)
	if err != nil {
		return nil, err
	}
	return yaml.Parse(string(y))
}
