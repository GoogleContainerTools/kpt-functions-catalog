package transformer

import (
	"encoding/json"
	"fmt"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/yaml"
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
		return false, nil
	}
	if err := transformer.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, nil
	}

	rl.Results = append(rl.Results, transformer.Results...)
	// TODO: another way to pass result.
	return true, nil
}

type LabelTransformer struct {
	// Desired labels
	NewLabels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalLabelFields instead.
	FieldSpecs []FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalLabelFields is used to specify additional fields to add labels. TODO: deprecated soon
	AdditionalLabelFields []FieldSpec `json:"additionalLabelFields,omitempty" yaml:"additionalLabelFields,omitempty"`
	// Results is used to track labels that have been applied
	Results fn.Results
}

// Config parse the functionConfig kubeObject to the fields in the LabelTransformer
func (p *LabelTransformer) Config(o *fn.KubeObject) error {
	// parse labels to NewLabels
	switch {
	case o.IsEmpty():
		return fmt.Errorf("FunctionConfig is missing. Expect `ConfigMap` or `SetLabel`")
	case o.IsGVK("", "v1", "ConfigMap"):
		p.NewLabels = o.NestedStringMapOrDie("data")
		if len(p.NewLabels) == 0 {
			return fmt.Errorf("`data` should not be empty")
		}
	case o.IsGVK(fnConfigGroup, fnConfigAPIVersion, legacyFnConfigKind):
		fallthrough
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
	// add additional fields
	if o.IsGVK(fnConfigGroup, fnConfigAPIVersion, fnConfigKind) {
		var add []FieldSpec
		if _, err := o.Get(&add, "additionalLabelFields"); err != nil {
			return err
		}
		p.AdditionalLabelFields = append(p.AdditionalLabelFields, add...)
	}
	return nil
}

func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	for _, o := range objects {
		for _, sp := range p.AdditionalLabelFields {
			if (sp.Group == "" && sp.Version == "" && sp.Kind == "") || o.IsGVK(sp.Group, sp.Version, sp.Kind) {
				// generate msg
				res, _ := json.Marshal(p.NewLabels)
				newResult := fn.Result{
					Message:     "set labels: " + string(res),
					Severity:    "",
					ResourceRef: nil,
					Field: &fn.Field{
						Path:          sp.Path,
						CurrentValue:  nil, // values to be updated in setLabel()
						ProposedValue: nil, // values to be updated in setLabel()
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
	for k, v := range newLabels {
		newPath := append(basePath, k)
		_, exist, err := o.NestedString(newPath...)
		if err != nil {
			return err
		}
		if exist || create {
			if err = o.SetNestedString(v, newPath...); err != nil {
				return err
			}
		}
	}
	return nil

}

func (p *LabelTransformer) addDefaultLabelFields() error {
	var defaultFieldSpecs []FieldSpec
	err := yaml.Unmarshal([]byte(commonLabelFieldSpecs), &defaultFieldSpecs)
	if err != nil {
		return err
	}
	p.AdditionalLabelFields = append(p.AdditionalLabelFields, defaultFieldSpecs...)
	return nil
}
