package transformer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

type FieldPath []string

type GVK struct {
	group   string
	version string
	kind    string
}

// FieldSpec stores information about how to modify a specific label
type FieldSpec struct {
	Gvk       GVK
	FieldPath FieldPath
	// TODO: should support user configurable field for labels
	CreateIfNotPresent bool
}

type FieldSpecs []FieldSpec

// LabelTransformer supports the set-labels workflow, it uses Config to parse functionConfig, Transform to change the labels
type LabelTransformer struct {
	// NewLabels is the desired labels
	NewLabels map[string]string
	// FieldSpecs stores default label fields
	FieldSpecs []FieldSpec
	// Results logs the changes to the KRM resource labels
	Results fn.Results
}

// SetLabels perform the whole set labels operation according to given resourcelist
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
	return true, nil
}

// Config parse the functionConfig kubeObject to the fields in the LabelTransformer
func (p *LabelTransformer) Config(functionConfig *fn.KubeObject) error {
	// parse labels to NewLabels
	switch {
	case functionConfig.IsEmpty():
		return fmt.Errorf("Config is Empty, failed to configure function: `functionConfig` must be either a `ConfigMap` or `SetLabels`")
	case functionConfig.IsGVK("", "v1", "ConfigMap"):
		p.NewLabels = functionConfig.NestedStringMapOrDie("data")
	case functionConfig.IsGVK(fnConfigGroup, fnConfigAPIVersion, fnConfigKind):
		if _, exist, err := functionConfig.NestedSlice(fnDeprecateField); exist || err != nil {
			return fmt.Errorf("`additionalLabelFields` has been deprecated")
		}
		p.NewLabels = functionConfig.NestedStringMapOrDie("labels")
		if len(p.NewLabels) == 0 {
			return fmt.Errorf("failed to configure function: input label list cannot be empty, required valid `labels` field")
		}
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap` with correct formatting",
			functionConfig.GetKind(), functionConfig.GetAPIVersion(), fnConfigKind)
	}
	return nil
}

// setLabelsInSpecs sets labels according to the generated common label
func (p *LabelTransformer) setLabelsInSpecs(o *fn.KubeObject) error {
	for _, spec := range CommonSpecs {
		if o.IsGVK(spec.Gvk.group, spec.Gvk.version, spec.Gvk.kind) {
			updatedLabels, err := updateLabels(&o.SubObject, spec.FieldPath, p.NewLabels, spec.CreateIfNotPresent)
			if err != nil {
				return err
			}
			p.LogResult(o, spec.FieldPath, updatedLabels)
		}
	}
	return nil
}

func IsLocalConfig(o *fn.KubeObject) bool {
	return o.IsLocalConfig()
}

// Transform updates the labels in the right path using GVK filter and other configurable fields
func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	// check if local config
	objects = objects.WhereNot(IsLocalConfig)
	for _, o := range objects {
		// this label need to set for all GVK
		metaLabelsPath := FieldPath{"metadata", "labels"}
		updatedLabels, err := updateLabels(&o.SubObject, metaLabelsPath, p.NewLabels, true)
		if err != nil {
			return err
		}
		p.LogResult(o, metaLabelsPath, updatedLabels)
		// set other common labels according to specific GVK
		err = p.setLabelsInSpecs(o)
		if err != nil {
			return err
		}
		// handle special case with slice
		if o.IsGVK("apps", "", "StatefulSet") {
			if o.GetMap("spec") != nil {
				for _, vctObj := range o.GetMap("spec").GetSlice("volumeClaimTemplates") {
					updatedLabels, err = updateLabels(vctObj, metaLabelsPath, p.NewLabels, true)
					if err != nil {
						return err
					}
					p.LogResult(o, FieldPath{"spec", "volumeClaimTemplates[]", "metadata", "labels"}, updatedLabels)
				}
			}
		}
	}
	return nil
}

// LogResult logs the KRM resource that has the labels changed
func (p *LabelTransformer) LogResult(o *fn.KubeObject, path []string, labels map[string]string) {
	res, _ := json.Marshal(labels)
	newResult := fn.Result{
		Message:     "set labels: " + string(res),
		Severity:    "",
		ResourceRef: nil,
		Field: &fn.Field{
			Path:          strings.Join(path, "."),
			CurrentValue:  nil,
			ProposedValue: nil,
		},
		File: &fn.File{
			Path:  o.PathAnnotation(),
			Index: o.IndexAnnotation(),
		},
		Tags: nil,
	}
	p.Results = append(p.Results, &newResult)
}

// updateLabels the update process for each label, sort the keys to preserve sequence, return if the update was performed and potential error
func updateLabels(o *fn.SubObject, labelPath FieldPath, newLabels map[string]string, create bool) (map[string]string, error) {
	keys := make([]string, 0)
	for k := range newLabels {
		keys = append(keys, k)
	}
	updatedLabels := make(map[string]string)
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := newLabels[key]
		newPath := append(labelPath, key)
		oldValue, exist, err := o.NestedString(newPath...)
		if err != nil {
			return nil, err
		}
		//TODO: should support user configurable field for labels
		if (exist && oldValue != val) || (!exist && create) {
			if err = o.SetNestedString(val, newPath...); err != nil {
				return nil, err
			}
			updatedLabels[key] = val
		}
	}
	return updatedLabels, nil
}
