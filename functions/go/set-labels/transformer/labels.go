package transformer

import (
	"encoding/json"
	"fmt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sort"
	"strings"
)

// the similar struct esxits in resid.GVK, but there is no function to create an GVK struct without using kyaml
type FieldSpec struct {
	Group              string `json:"group,omitempty" yaml:"group,omitempty"`
	Version            string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind               string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"` // seperated by /
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
}

func SetLabels(rl *fn.ResourceList) (bool, error) {
	transformer := LabelTransformer{}
	if err := transformer.Config(rl.FunctionConfig); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, err
	}
	if err := transformer.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, err
	}

	rl.Results = append(rl.Results, transformer.Results...)
	return true, nil
}

type LabelTransformer struct {
	// Desired labels
	NewLabels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalLabelFields instead.
	FieldSpecs []FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// Results is used to track labels that have been applied
	Results fn.Results
}

// Config parse the functionConfig kubeObject to the fields in the LabelTransformer
func (p *LabelTransformer) Config(o *fn.KubeObject) error {
	// parse labels to NewLabels
	switch {
	case o.IsEmpty():
		return fmt.Errorf("failed to configure function: `functionConfig` must be either a `ConfigMap` or `SetLabels`")
	case o.IsGVK("", "v1", "ConfigMap"):
		p.NewLabels = o.NestedStringMapOrDie("data")
	case o.IsGVK(fnConfigGroup, fnConfigAPIVersion, fnConfigKind):
		p.NewLabels = o.NestedStringMapOrDie("labels")
		if len(p.NewLabels) == 0 {
			return fmt.Errorf("failed to configure function: input label list cannot be empty")
		}
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap`",
			o.GetKind(), o.GetAPIVersion(), fnConfigKind)
	}
	// add default fields
	if err := p.addDefaultLabelFields(); err != nil {
		return err
	}
	return nil
}

func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	for _, o := range objects {
		for _, sp := range p.FieldSpecs {
			if (sp.Group == "" && sp.Version == "" && sp.Kind == "") || o.IsGVK(sp.Group, sp.Version, sp.Kind) {
				// generate msg
				res, _ := json.Marshal(p.NewLabels)
				newResult := fn.Result{
					Message:     "set labels: " + string(res),
					Severity:    "",
					ResourceRef: nil,
					Field: &fn.Field{
						Path:          sp.Path,
						CurrentValue:  nil,
						ProposedValue: nil,
					},
					File: &fn.File{
						Path:  o.PathAnnotation(),
						Index: o.IndexAnnotation(),
					},
					Tags: nil,
				}
				err := updateLabels(o, sp.Path, p.NewLabels, sp.CreateIfNotPresent)
				if err != nil {
					return err
				}
				p.Results = append(p.Results, &newResult)
			}
		}
	}
	return nil
}

func updateLabels(o *fn.KubeObject, fieldPath string, newLabels map[string]string, create bool) error {
	//TODO: should support user configurable field for labels
	basePath := strings.Split(fieldPath, "/")
	keys := make([]string, 0)
	for k := range newLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := newLabels[key]
		newPath := append(basePath, key)
		_, exist, err := o.NestedString(newPath...)
		if err != nil {
			return err
		}
		if exist || create {
			if err = o.SetNestedString(val, newPath...); err != nil {
				return err
			}
		}
	}
	return nil

}

func (p *LabelTransformer) addDefaultLabelFields() error {
	var defaultFieldSpecs []FieldSpec
	err := json.Unmarshal([]byte(commonLabelFieldSpecs), &defaultFieldSpecs)
	if err != nil {
		return err
	}
	p.FieldSpecs = append(p.FieldSpecs, defaultFieldSpecs...)
	return nil
}
