package transformer

import (
	"fmt"
	"sort"

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
	// ResultCount logs the total count of each labels change
	ResultCount map[string]int
}

// NewTransformer is the constructor for labelTransformer
func NewTransformer() *LabelTransformer {
	transformer := LabelTransformer{}
	resultCount := make(map[string]int)
	transformer.ResultCount = resultCount
	return &transformer
}

// SetLabels perform the whole set labels operation according to given resourcelist
func SetLabels(rl *fn.ResourceList) (bool, error) {
	transformer := NewTransformer()
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
	case functionConfig.IsGVK(fn.KptFunctionGroup, fn.KptFunctionVersion, FnConfigKind):
		if _, exist, err := functionConfig.NestedSlice(fnDeprecateField); exist || err != nil {
			return fmt.Errorf("`additionalLabelFields` has been deprecated")
		}
		p.NewLabels = functionConfig.NestedStringMapOrDie("labels")
		if len(p.NewLabels) == 0 {
			return fmt.Errorf("failed to configure function: input label list cannot be empty, required valid `labels` field")
		}
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap` with correct formatting",
			functionConfig.GetKind(), functionConfig.GetAPIVersion(), FnConfigKind)
	}
	return nil
}

// setLabelsInSpecs sets labels according to the generated common label
func (p *LabelTransformer) setCommonSpecLabels(o *fn.KubeObject) error {
	for _, spec := range CommonSpecs {
		if o.IsGVK(spec.Gvk.group, spec.Gvk.version, spec.Gvk.kind) {
			err := updateLabels(&o.SubObject, spec.FieldPath, p.NewLabels, spec.CreateIfNotPresent, p.ResultCount)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Transform updates the labels in the right path using GVK filter and other configurable fields
func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	for _, o := range objects.WhereNot(func(o *fn.KubeObject) bool { return o.IsLocalConfig() }) {
		// this label need to set for all GVK
		if err := p.setMetadataForAll(o); err != nil {
			return err
		}
		// set other common labels according to specific GVK
		if err := p.setCommonSpecLabels(o); err != nil {
			return err
		}
		// handle special cases when slices are involved
		if err := p.setLabelsInSlice(o); err != nil {
			return err
		}
		p.LogResult(o, p.ResultCount)
	}
	return nil
}

func (p *LabelTransformer) setMetadataForAll(o *fn.KubeObject) error {
	metaLabelsPath := FieldPath{"metadata", "labels"}
	err := updateLabels(&o.SubObject, metaLabelsPath, p.NewLabels, true, p.ResultCount)
	if err != nil {
		return err
	}
	return nil
}

// setLabelsInSlice handles the resources that contain slice type
func (p *LabelTransformer) setLabelsInSlice(o *fn.KubeObject) error {
	// handle resources that have podSpec struct
	if err := p.podSpecCheckAndUpdate(o); err != nil {
		return err
	}
	// handle other special case resources
	if err := p.specialCasesCheckAndUpdate(o); err != nil {
		return err
	}
	return nil
}

// podSpecCheckAndUpdate updates labels path inside podSpec struct
func (p *LabelTransformer) podSpecCheckAndUpdate(o *fn.KubeObject) error {
	if o.IsGVK("", "v1", "ReplicationController") ||
		o.IsGVK("", "", "Deployment") ||
		o.IsGVK("", "", "ReplicaSet") ||
		o.IsGVK("", "", "DaemonSet") ||
		o.IsGVK("apps", "", "StatefulSet") ||
		o.IsGVK("batch", "", "Job") {
		_, exist, _ := o.NestedString(FieldPath{"spec", "template", "spec"}...)
		if exist {
			podSpecObj := o.GetMap("spec").GetMap("template").GetMap("spec")
			if err := p.processPodSpec(podSpecObj, o); err != nil {
				return err
			}
		}
	}
	return nil
}

// processPodSpec takes in podSpec object and parse its path, parent kubeObject is also passed in for logging
func (p *LabelTransformer) processPodSpec(o *fn.SubObject, parentO *fn.KubeObject) error {
	labelSelector := FieldPath{"labelSelector", "matchLabels"}

	_, exist, _ := o.NestedSlice("topologySpreadConstraints")
	if exist {
		for _, obj := range o.GetSlice("topologySpreadConstraints") {
			err := updateLabels(obj, labelSelector, p.NewLabels, false, p.ResultCount)
			if err != nil {
				return err
			}
		}
	}

	subObj := o.GetMap("affinity")
	if subObj != nil {
		for _, aff := range []string{"podAffinity", "podAntiAffinity"} {
			ssubObj := subObj.GetMap(aff)
			if ssubObj != nil {
				for _, obj := range subObj.GetSlice("preferredDuringSchedulingIgnoredDuringExecution") {
					nxtObj := obj.GetMap("podAffinityTerm")
					if nxtObj != nil {
						err := updateLabels(nxtObj, labelSelector, p.NewLabels, false, p.ResultCount)
						if err != nil {
							return err
						}
					}

				}
				for _, obj := range subObj.GetSlice("requiredDuringSchedulingIgnoredDuringExecution") {
					err := updateLabels(obj, labelSelector, p.NewLabels, false, p.ResultCount)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// specialCasesCheckAndUpdate updates other paths that contain labels
func (p *LabelTransformer) specialCasesCheckAndUpdate(o *fn.KubeObject) error {
	metaLabelPath := FieldPath{"metadata", "labels"}
	if o.IsGVK("apps", "", "StatefulSet") {
		if o.GetMap("spec") != nil {
			for _, vctObj := range o.GetMap("spec").GetSlice("volumeClaimTemplates") {
				err := updateLabels(vctObj, metaLabelPath, p.NewLabels, false, p.ResultCount)
				if err != nil {
					return err
				}
			}
		}
	}

	if o.IsGVK("batch", "", "CronJob") {
		_, exist, _ := o.NestedString(FieldPath{"spec", "jobTemplate", "spec", "template", "spec"}...)
		if exist {
			podSpecObj := o.GetMap("spec").GetMap("jobTemplate").GetMap("spec").GetMap("template").GetMap("spec")
			if err := p.processPodSpec(podSpecObj, o); err != nil {
				return err
			}
		}
	}

	if o.IsGVK("networking.k8s.io", "", "NetworkPolicy") {
		podSelector := FieldPath{"podSelector", "matchLabels"}
		spec := o.GetMap("spec")
		if spec != nil {
			for _, vecObj := range spec.GetSlice("ingress") {
				for _, nextVecObj := range vecObj.GetSlice("from") {
					err := updateLabels(nextVecObj, podSelector, p.NewLabels, false, p.ResultCount)
					if err != nil {
						return err
					}

				}
			}
			for _, vecObj := range spec.GetSlice("egress") {
				for _, nextVecObj := range vecObj.GetSlice("to") {
					err := updateLabels(nextVecObj, podSelector, p.NewLabels, false, p.ResultCount)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// LogResult logs the KRM resource that has the labels changed
func (p *LabelTransformer) LogResult(o *fn.KubeObject, labelCount map[string]int) {
	// no labels get updated, no log
	if len(labelCount) == 0 {
		return
	}
	keys := make([]string, 0)
	for k := range labelCount {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		element := labelCount[key]
		msg := fmt.Sprintf("set labels {%v} for %v times", key, element)
		newResult := fn.Result{
			Message:     msg,
			Severity:    "",
			ResourceRef: nil,
			Field:       nil,
			File: &fn.File{
				Path:  o.PathAnnotation(),
				Index: o.IndexAnnotation(),
			},
			Tags: nil,
		}
		p.Results = append(p.Results, &newResult)
	}
}

// updateLabels the update process for each label, sort the keys to preserve sequence, return if the update was performed and potential error
func updateLabels(o *fn.SubObject, labelPath FieldPath, newLabels map[string]string, create bool, updatedLabelsCount map[string]int) error {
	keys := make([]string, 0)
	for k := range newLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := newLabels[key]
		newPath := append(labelPath, key)
		oldValue, exist, err := o.NestedString(newPath...)
		if err != nil {
			return err
		}
		//TODO: should support user configurable field for labels
		if (exist && oldValue != val) || (!exist && create) {
			if err = o.SetNestedString(val, newPath...); err != nil {
				return err
			}
			recordLabel := fmt.Sprintf("%v : %v", key, val)
			updatedLabelsCount[recordLabel] += 1
		}
	}
	return nil
}
