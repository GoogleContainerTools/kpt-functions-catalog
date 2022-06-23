package transformer

import (
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
	return true, nil
}

type LabelTransformer struct {
	// Desired labels
	NewLabels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalLabelFields instead.
	FieldSpecs []FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalLabelFields is used to specify additional fields to add labels.
	AdditionalLabelFields []FieldSpec `json:"additionalLabelFields,omitempty" yaml:"additionalLabelFields,omitempty"`
	// Results is used to track labels that have been applied
	Results fn.Results
}

// Config parse the functionConfig kubeObject to the fields in the LabelTransformer
func (p *LabelTransformer) Config(o *fn.KubeObject) error {
	// parse labels to NewLabels
	switch {
	case o.IsEmpty():
		return fmt.Errorf("FunctionConfig is missing. Expect `ConfigMap` or `SetNamespace`")
	case o.IsGVK("", "", "ConfigMap"):
		p.NewLabels = o.NestedStringMapOrDie("data")
		if len(p.NewLabels) == 0 {
			return fmt.Errorf("`data` should not be empty")
		}
	case o.IsGVK(fnConfigGroup, fnConfigAPIVersion, legacyFnConfigKind):
		fallthrough
	case o.IsGVK(fnConfigGroup, fnConfigAPIVersion, fnConfigKind):
		p.NewLabels = o.NestedStringMapOrDie("labels")
		if len(p.NewLabels) == 0 {
			return fmt.Errorf("`labels` should not be empty")
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
		arr, exist, err := o.NestedSlice("additionalLabelFields")
		if err != nil {
			return fmt.Errorf("`additionalLabelFields` format is wrong")
		}
		if exist {
			var addFields []FieldSpec
			for _, sub := range arr {
				addFields = append(addFields, FieldSpec{ //TODO: initialize in kyaml
					sub.GetString("group"),
					sub.GetString("version"), sub.GetString("kind"),
					sub.GetString("path"),
					sub.GetBool("create"),
				})
			}
			p.AdditionalLabelFields = append(p.AdditionalLabelFields, addFields...)
		}
	}
	return nil
}

func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	for _, o := range objects {
		for _, sp := range p.AdditionalLabelFields {
			if (sp.Group == "" && sp.Version == "" && sp.Kind == "") || o.IsGVK(sp.Group, sp.Version, sp.Kind) {
				newResult := fn.Result{
					Message:     "Replace labels",
					Severity:    "INFO",
					ResourceRef: nil, // TODO: initialize in kyaml
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
				err := updateLabels(o, sp.Path, p.NewLabels, sp.CreateIfNotPresent, &newResult)
				if err != nil {
					return err
				}
				p.Results = append(p.Results, &newResult)
			}
		}
	}
	return nil
}

// TODO: Is there any helper function insdie fn package to help with copy map
func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	newMap := make(map[string]string)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func updateLabels(o *fn.KubeObject, fieldPath string, newLabels map[string]string, create bool, result *fn.Result) error {
	paths := strings.Split(fieldPath, "/")

	oldLabels, exist, err := o.NestedStringMap(paths...)
	if err != nil {
		return err
	}
	if !exist {
		oldLabels = make(map[string]string)
	}
	result.Field.CurrentValue = copyStringMap(oldLabels)
	replaceLabels(oldLabels, newLabels, create)
	result.Field.ProposedValue = copyStringMap(oldLabels)
	o.SetNestedStringMapOrDie(oldLabels, paths...)
	return nil
}

// replaceLabels replace old labels map with new labels map according to create
// oldLabels must not be nil
func replaceLabels(oldLabels map[string]string, newLabels map[string]string, create bool) {
	for k, v := range newLabels {
		_, exist := oldLabels[k]
		if create || exist {
			oldLabels[k] = v
		}
	}
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
