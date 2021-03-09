package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/starlark"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	fnConfigGroup      = "kpt.dev"
	fnConfigVersion    = "v1beta1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "StarlarkFunction"
)

type StarlarkFunction struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Source is a required field for providing a starlark script.
	Source Source `json:"source" yaml:"source"`
	// KeyValues is a convenient way to pass in arbitrary key value pairs.
	KeyValues map[string]string `json:"keyValues,omitempty" yaml:"keyValues,omitempty"`
}

// Source contains an untagged union, only one field can be set.
type Source struct {
	// Inline is a starlark script in string format.
	Inline string `json:"inline,omitempty" yaml:"inline,omitempty"`
	// Path is the path to a starlark script.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// URL is the url of a remote starlark script.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}

func (sf *StarlarkFunction) Validate() error {
	if sf.APIVersion != fnConfigAPIVersion {
		return fmt.Errorf("apiVersion is expected to be: %v", fnConfigAPIVersion)
	}
	if sf.Kind != fnConfigKind {
		return fmt.Errorf("kind is expected to be: %v", fnConfigKind)
	}

	if sf.ObjectMeta.Name == "" {
		return fmt.Errorf("name is required in starlark function config")
	}

	if (sf.Source.Inline != "" && sf.Source.Path != "") ||
		(sf.Source.Path != "" && sf.Source.URL != "") ||
		(sf.Source.Inline != "" && sf.Source.URL != "") {
		return fmt.Errorf("only one of inline, path and url can be set")
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
		Program: sf.Source.Inline,
		Path:    sf.Source.Path,
		URL:     sf.Source.URL,
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
