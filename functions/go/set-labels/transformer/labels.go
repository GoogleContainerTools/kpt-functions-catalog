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

// Transform updates the labels in the right path using GVK filter and other configurable fields
func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	for _, o := range objects.WhereNot(func(o *fn.KubeObject) bool { return o.IsLocalConfig() }) {
		err := func() error {
			// set meta labels
			if err := p.setObjectMeta(o); err != nil {
				return err
			}
			// set selector
			if err := p.setSelector(o); err != nil {
				return err
			}
			// set PodTemplateSpec
			if err := p.setPodTemplateSpec(o); err != nil {
				return err
			}
			// set other corner cases
			if err := p.setVolumeClaimTemplates(o); err != nil {
				return err
			}
			if err := p.setJobTemplateSpecMeta(o); err != nil {
				return err
			}
			if err := p.setNetworkPolicyRule(o); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
		p.LogResult(o, p.ResultCount)
	}
	return nil
}

func (p *LabelTransformer) setNetworkPolicyRule(o *fn.KubeObject) error {
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
	return nil
}

func hasJobTemplateSpec(o *fn.KubeObject) bool {
	return o.IsGVK("batch", "", "CronJob")
}

func (p *LabelTransformer) setJobTemplateSpecMeta(o *fn.KubeObject) error {
	if hasJobTemplateSpec(o) {
		// set objectMeta
		fieldPath := FieldPath{"spec", "jobTemplate", "metadata", "labels"}
		if err := updateLabels(&o.SubObject, fieldPath, p.NewLabels, true, p.ResultCount); err != nil {
			return err
		}
	}
	return nil
}

func (p *LabelTransformer) setJobSpecObjectMeta(o *fn.KubeObject) error {
	// set podTemplateSpec objectMeta
	specfieldPath := FieldPath{"spec", "jobTemplate", "spec", "template", "metadata", "labels"}
	if err := updateLabels(&o.SubObject, specfieldPath, p.NewLabels, true, p.ResultCount); err != nil {
		return err
	}
	return nil
}

func (p *LabelTransformer) setJobPodSpec(o *fn.KubeObject) error {
	// set podTemplateSpec affinity
	if hasJobPodSpec(o) {
		podSpecObj := o.GetMap("spec").GetMap("jobTemplate").GetMap("spec").GetMap("template").GetMap("spec")
		if err := p.setPodSpec(podSpecObj); err != nil {
			return err
		}
	}
	return nil
}

func hasJobPodSpec(o *fn.KubeObject) bool {
	if cronJobSpec := o.GetMap("spec"); cronJobSpec != nil {
		if jobTemplateSpec := cronJobSpec.GetMap("jobTemplate"); jobTemplateSpec != nil {
			if jobSpec := jobTemplateSpec.GetMap("spec"); jobSpec != nil {
				if podTemplateSpec := jobSpec.GetMap("template"); podTemplateSpec != nil {
					if podSpec := podTemplateSpec.GetMap("spec"); podSpec != nil {
						return true
					}
				}
			}
		}
	}
	return false
}

func (p *LabelTransformer) setSelector(o *fn.KubeObject) error {
	if hasSpecSelector(o) {
		fieldPath := FieldPath{"spec", "selector"}
		if err := updateLabels(&o.SubObject, fieldPath, p.NewLabels, true, p.ResultCount); err != nil {
			return err
		}
	}
	if found, create := hasLabelSelector(o); found {
		fieldPath := FieldPath{"spec", "selector", "matchLabels"}
		if err := updateLabels(&o.SubObject, fieldPath, p.NewLabels, create, p.ResultCount); err != nil {
			return err
		}
	}
	if hasJobTemplateSpec(o) {
		fieldPath := FieldPath{"spec", "jobTemplate", "spec", "selector", "matchLabels"}
		if err := updateLabels(&o.SubObject, fieldPath, p.NewLabels, false, p.ResultCount); err != nil {
			return err
		}
	}
	if hasNetworkPolicySpec(o) {
		fieldPath := FieldPath{"spec", "podSelector", "matchLabels"}
		if err := updateLabels(&o.SubObject, fieldPath, p.NewLabels, false, p.ResultCount); err != nil {
			return err
		}
	}
	return nil
}

func hasNetworkPolicySpec(o *fn.KubeObject) bool {
	return o.IsGVK("networking.k8s.io", "", "NetworkPolicy")
}

func hasSpecSelector(o *fn.KubeObject) bool {
	if o.IsGVK("", "v1", "Service") || o.IsGVK("", "v1", "ReplicationController") {
		return true
	}
	return false
}

// hasLabelSelector return (if the resource has LabelSelector, if the LabelSelector need to be created if not exist)
func hasLabelSelector(o *fn.KubeObject) (bool, bool) {
	if o.IsGVK("", "", "Deployment") || o.IsGVK("", "", "ReplicaSet") || o.IsGVK("", "", "DaemonSet") || o.IsGVK("apps", "", "StatefulSet") {
		return true, true
	}
	if o.IsGVK("batch", "", "Job") || o.IsGVK("policy", "", "PodDisruptionBudget") {
		return true, false
	}
	return false, false
}

func hasPodTemplateSpec(o *fn.KubeObject) bool {
	if o.IsGVK("", "v1", "ReplicationController") ||
		o.IsGVK("", "", "Deployment") ||
		o.IsGVK("", "", "ReplicaSet") ||
		o.IsGVK("", "", "DaemonSet") ||
		o.IsGVK("apps", "", "StatefulSet") ||
		o.IsGVK("batch", "", "Job") {
		return true
	}
	return false
}

func hasVolumeClaimTemplates(o *fn.KubeObject) bool {
	return o.IsGVK("apps", "", "StatefulSet")
}

func (p *LabelTransformer) setVolumeClaimTemplates(o *fn.KubeObject) error {
	if hasVolumeClaimTemplates(o) {
		metaLabelPath := FieldPath{"metadata", "labels"}
		if o.GetMap("spec") != nil {
			for _, vctObj := range o.GetMap("spec").GetSlice("volumeClaimTemplates") {
				err := updateLabels(vctObj, metaLabelPath, p.NewLabels, true, p.ResultCount)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *LabelTransformer) setPodTemplateSpec(o *fn.KubeObject) error {
	if hasPodTemplateSpec(o) {
		// set objectMeta
		if err := p.setSpecObjectMeta(o); err != nil {
			return err
		}
		// set affinity
		if hasPodSpec(o) {
			if err := p.setPodSpec(o.GetMap("spec").GetMap("template").GetMap("spec")); err != nil {
				return err
			}
		}
	}
	// PodTemplateSpec can also be in JobTemplateSpec
	if hasJobTemplateSpec(o) {
		if err := p.setJobSpecObjectMeta(o); err != nil {
			return err
		}
		if err := p.setJobPodSpec(o); err != nil {
			return err
		}
	}
	return nil
}

func hasPodSpec(o *fn.KubeObject) bool {
	if spec := o.GetMap("spec"); spec != nil {
		if template := spec.GetMap("template"); template != nil {
			if podSpec := template.GetMap("spec"); podSpec != nil {
				return true
			}
		}
	}
	return false
}

func (p *LabelTransformer) setPodSpec(podSpec *fn.SubObject) error {
	labelSelector := FieldPath{"labelSelector", "matchLabels"}

	_, exist, _ := podSpec.NestedSlice("topologySpreadConstraints")
	if exist {
		for _, obj := range podSpec.GetSlice("topologySpreadConstraints") {
			err := updateLabels(obj, labelSelector, p.NewLabels, false, p.ResultCount)
			if err != nil {
				return err
			}
		}
	}

	subObj := podSpec.GetMap("affinity")
	if subObj != nil {
		for _, aff := range []string{"podAffinity", "podAntiAffinity"} {
			podAff := subObj.GetMap(aff)
			if podAff != nil {
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

// setSpecObjectMeta takes in subObject and check all its field for ObjectMeta, create key value if not existed
func (p *LabelTransformer) setSpecObjectMeta(o *fn.KubeObject) error {
	metaLabelsPath := FieldPath{"spec", "template", "metadata", "labels"}
	err := updateLabels(&o.SubObject, metaLabelsPath, p.NewLabels, true, p.ResultCount)
	if err != nil {
		return err
	}
	return nil
}

func (p *LabelTransformer) setObjectMeta(o *fn.KubeObject) error {
	// all resources' ObjectMeta labels need to be updated
	metaLabelsPath := FieldPath{"metadata", "labels"}
	err := updateLabels(&o.SubObject, metaLabelsPath, p.NewLabels, true, p.ResultCount)
	if err != nil {
		return err
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
