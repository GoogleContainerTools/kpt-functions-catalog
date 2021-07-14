package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/starlark/third_party/sigs.k8s.io/kustomize/kyaml/fn/runtime/starlark"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	starlarkRunGroup                   = "fn.kpt.dev"
	starlarkRunVersion                 = "v1alpha1"
	starlarkRunAPIVersion              = starlarkRunGroup + "/" + starlarkRunVersion
	starlarkRunKind       fnConfigKind = "StarlarkRun"

	configMapApiVersion              = "v1"
	configMapKind       fnConfigKind = "ConfigMap"

	sourceKey = "source"

	defaultProgramName = "stalark-function-run"
)

type fnConfigKind string

type StarlarkFnConfig struct {
	kind        fnConfigKind
	starlarkRun *StarlarkRun
	configMap   *corev1.ConfigMap
}

func (sfc *StarlarkFnConfig) GetName() string {
	switch sfc.kind {
	case configMapKind:
		return sfc.configMap.Name
	case starlarkRunKind:
		return sfc.starlarkRun.Name
	default:
		return ""
	}
}

func (sfc *StarlarkFnConfig) GetSource() string {
	switch sfc.kind {
	case configMapKind:
		return sfc.configMap.Data["source"]
	case starlarkRunKind:
		return sfc.starlarkRun.Source
	default:
		return ""
	}
}

type StarlarkRun struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Source is a required field for providing a starlark script inline.
	Source string `json:"source" yaml:"source"`
	// Params are the parameters in key-value pairs format.
	Params map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

var _ framework.Defaulter = &StarlarkFnConfig{}

func (sfc *StarlarkFnConfig) Default() error {
	switch sfc.kind {
	case configMapKind:
		if sfc.configMap.Name == "" {
			sfc.configMap.Name = defaultProgramName
		}
		return nil
	case starlarkRunKind:
		if sfc.starlarkRun.Name == "" {
			sfc.starlarkRun.Name = defaultProgramName
		}
		return nil
	default:
		return fmt.Errorf("unknown `functionConfig` kind: %v", sfc.kind)
	}
}

var _ framework.Validator = &StarlarkFnConfig{}

func (sfc *StarlarkFnConfig) Validate() error {
	switch sfc.kind {
	case configMapKind:
		return validateConfigMap(sfc.configMap)
	case starlarkRunKind:
		return validateStarlarkRun(sfc.starlarkRun)
	default:
		return fmt.Errorf("unknown `functionConfig` kind: %v", sfc.kind)
	}
}

func validateConfigMap(cm *corev1.ConfigMap) error {
	if cm.APIVersion != "v1" {
		return fmt.Errorf("`apiVersion` must be %q when using `ConfigMap` as the `functionConfig`, but got %q", "v1", cm.APIVersion)
	}
	if cm.Kind != string(configMapKind) {
		return fmt.Errorf("`kind` must be: %q when using `ConfigMap` as the `functionConfig`, but got %q", configMapKind, cm.Kind)
	}
	if cm.Data == nil {
		return fmt.Errorf("`data.source` must not be empty in `ConfigMap`")
	}
	if _, found := cm.Data[sourceKey]; !found {
		return fmt.Errorf("`data.source` must not be empty in `ConfigMap`")
	}
	return nil
}

func validateStarlarkRun(sr *StarlarkRun) error {
	if sr.APIVersion != starlarkRunAPIVersion {
		return fmt.Errorf("`apiVersion` must be %q when using `StarlarkRun` as the `functionConfig`, but got %q", starlarkRunAPIVersion, sr.APIVersion)
	}
	if sr.Kind != string(starlarkRunKind) {
		return fmt.Errorf("`kind` must be: %q when using `StarlarkRun` as the `functionConfig`, but got %q", starlarkRunKind, sr.Kind)
	}
	if sr.Source == "" {
		return fmt.Errorf("`source` in `StarlarkRun` must not be empty")
	}
	return nil
}

func (sfc *StarlarkFnConfig) Transform(rl *framework.ResourceList) error {
	var err error
	starFltr := &starlark.SimpleFilter{
		Name:           sfc.GetName(),
		Program:        sfc.GetSource(),
		FunctionConfig: rl.FunctionConfig,
	}
	rl.Items, err = starFltr.Filter(rl.Items)
	return err
}

func (sfc *StarlarkFnConfig) UnmarshalYAML(value *yaml.Node) error {
	rn := yaml.NewRNode(value)
	meta, err := rn.GetMeta()
	if err != nil {
		return err
	}
	s, err := rn.String()
	if err != nil {
		return err
	}

	switch {
	case meta.APIVersion == configMapApiVersion && meta.Kind == string(configMapKind):
		cm := corev1.ConfigMap{}
		if err = k8syaml.Unmarshal([]byte(s), &cm); err != nil {
			return fmt.Errorf("unable to unmarshal the `ConfigMap`: %w", err)
		}
		sfc.kind = configMapKind
		sfc.configMap = &cm
	case meta.APIVersion == starlarkRunAPIVersion && meta.Kind == string(starlarkRunKind):
		sr := StarlarkRun{}
		if err = k8syaml.Unmarshal([]byte(s), &sr); err != nil {
			return fmt.Errorf("unable to unmarshal the `StarlarkRun`: %w", err)
		}
		sfc.kind = starlarkRunKind
		sfc.starlarkRun = &sr
	default:
		return fmt.Errorf("`functionConfig` must be either %v or %v, but we got: %v",
			schema.FromAPIVersionAndKind(configMapApiVersion, string(configMapKind)).String(),
			schema.FromAPIVersionAndKind(starlarkRunAPIVersion, string(starlarkRunKind)).String(),
			schema.FromAPIVersionAndKind(meta.APIVersion, meta.Kind).String())
	}
	return nil
}
