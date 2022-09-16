package transformer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/types"
)

type FieldPath []string

// LabelTransformer supports the set-labels workflow, it uses Config to parse functionConfig, Transform to change the labels
type LabelTransformer struct {
	// NewLabels is the desired labels
	NewLabels map[string]string
	// Results logs the changes to the KRM resource labels
	Results fn.Results
	// ResultCount logs the total count of each labels change
	ResultCount map[string]int
}

// NewLabelTransformer is the constructor for labelTransformer
func NewLabelTransformer() *LabelTransformer {
	resultCount := make(map[string]int)
	return &LabelTransformer{
		ResultCount: resultCount,
	}
}

// SetLabels perform the whole set labels operation according to given resourcelist
func SetLabels(rl *fn.ResourceList) (bool, error) {
	transformer := NewLabelTransformer()
	if err := transformer.Config(rl.FunctionConfig, rl); err != nil {
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
func (p *LabelTransformer) Config(functionConfig *fn.KubeObject, rl *fn.ResourceList) error {
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
		labelsFrom, err := parseLabelsFrom(functionConfig, rl)
		if err != nil {
			return err
		}
		// merge labelsFrom with labelValues
		for k, v := range labelsFrom {
			p.NewLabels[k] = v
		}
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
	// using unit test and pass in empty string would provide a nil; an empty file in e2e would provide 0 object
	if objects.Len() == 0 || objects[0] == nil {
		newResult := fn.GeneralResult("no input resources", fn.Info)
		p.Results = append(p.Results, newResult)
		return nil
	}
	for _, o := range objects.WhereNot(func(o *fn.KubeObject) bool { return o.IsLocalConfig() }) {
		err := func() error {
			if err := p.setObjectMeta(o); err != nil {
				return err
			}
			if err := p.setSelector(o); err != nil {
				return err
			}
			if err := p.setPodTemplateSpec(o); err != nil {
				return err
			}
			if hasVolumeClaimTemplates(o) {
				if err := p.setVolumeClaimTemplates(o); err != nil {
					return err
				}
			}
			if hasJobTemplateSpec(o) {
				if err := p.setJobTemplateSpecMeta(o); err != nil {
					return err
				}
			}
			if hasNetworkPolicySpec(o) {
				if err := p.setNetworkPolicyRule(o); err != nil {
					return err
				}
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

// hasJobTemplateSpec check if the object is CronJob, this kind might have JobTemplateSpec, which contains label fieldPath
func hasJobTemplateSpec(o *fn.KubeObject) bool {
	return o.IsGVK("batch", "", "CronJob")
}

// setJobTemplateSpecMeta set MetaObject in struct JobTemplateSpec
func (p *LabelTransformer) setJobTemplateSpecMeta(o *fn.KubeObject) error {
	// set objectMeta
	fieldPath := FieldPath{"spec", "jobTemplate", "metadata", "labels"}
	if err := updateLabels(&o.SubObject, fieldPath, p.NewLabels, true, p.ResultCount); err != nil {
		return err
	}
	return nil
}

// setJobSpecObjectMeta set MetaObject in struct JobSpec
func (p *LabelTransformer) setJobSpecObjectMeta(o *fn.KubeObject) error {
	// set podTemplateSpec objectMeta
	specfieldPath := FieldPath{"spec", "jobTemplate", "spec", "template", "metadata", "labels"}
	if err := updateLabels(&o.SubObject, specfieldPath, p.NewLabels, true, p.ResultCount); err != nil {
		return err
	}
	return nil
}

// hasJobPodSpec check if the object contains struct JobPodSpec
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

// setJobPodSpec set labels path in PodSpec for kind CronJob
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

// hasNetworkPolicySpec checking if this kind is NetworkPolicy, it would contain struct NetworkPolicySpec
func hasNetworkPolicySpec(o *fn.KubeObject) bool {
	return o.IsGVK("networking.k8s.io", "", "NetworkPolicy")
}

// setNetworkPolicyRule set ingress and egress rules for kind NetworkPolicy
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

// hasSpecSelector check if the resource contains struct SpecSelector, kind Service and ReplicationController has it
func hasSpecSelector(o *fn.KubeObject) bool {
	if o.IsGVK("", "v1", "Service") || o.IsGVK("", "v1", "ReplicationController") {
		return true
	}
	return false
}

// setSelector set labels for all selectors, including spec selector map, spec selector LabelSelector, LabelSelector in JobTemplate, and podSelector in NetworkPolicy, and
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

// hasLabelSelector check if the resource contains struct LabelSelector, return (if the resource has LabelSelector, if the LabelSelector need to be created if not exist)
func hasLabelSelector(o *fn.KubeObject) (bool, bool) {
	if o.IsGVK("", "", "Deployment") || o.IsGVK("", "", "ReplicaSet") || o.IsGVK("", "", "DaemonSet") || o.IsGVK("apps", "", "StatefulSet") {
		return true, true
	}
	if o.IsGVK("batch", "", "Job") || o.IsGVK("policy", "", "PodDisruptionBudget") {
		return true, false
	}
	return false, false
}

// hasPodTemplateSpec check if the resource contains struct PodTemplateSpec, ReplicationController, Deployment, ReplicaSet, DaemonSet, StatefulSet, Job kind has it
func hasPodTemplateSpec(o *fn.KubeObject) bool {
	switch {
	case o.IsGVK("", "v1", "ReplicationController"):
		return true
	case o.IsGVK("", "", "Deployment"):
		return true
	case o.IsGVK("", "", "ReplicaSet"):
		return true
	case o.IsGVK("", "", "DaemonSet"):
		return true
	case o.IsGVK("apps", "", "StatefulSet"):
		return true
	case o.IsGVK("batch", "", "Job"):
		return true
	default:
		return false
	}
}

// setPodTemplateSpec set label path for PodTemplateSpec, both its ObjectMeta and its PodSpec
func (p *LabelTransformer) setPodTemplateSpec(o *fn.KubeObject) error {
	if hasPodTemplateSpec(o) {
		// set objectMeta
		if err := p.setSpecObjectMeta(o); err != nil {
			return err
		}
		// set podSpec
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

// hasVolumeClaimTemplates check if the resource contains struct VolumeClaimTemplates, kind StatefulSet has it
func hasVolumeClaimTemplates(o *fn.KubeObject) bool {
	return o.IsGVK("apps", "", "StatefulSet")
}

// setVolumeClaimTemplates set VolumeClaimTemplates label path
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

// hasPodSpec check if the resource has struct PodSpec
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

// setPodSpec set label path in PodSpec, that include path under topologySpreadConstraints and affinity
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

// setSpecObjectMeta takes in spec subObject and check its field for ObjectMeta, create key value if not existed
func (p *LabelTransformer) setSpecObjectMeta(o *fn.KubeObject) error {
	metaLabelsPath := FieldPath{"spec", "template", "metadata", "labels"}
	err := updateLabels(&o.SubObject, metaLabelsPath, p.NewLabels, true, p.ResultCount)
	if err != nil {
		return err
	}
	return nil
}

// setObjectMeta set ObjectMeta labels for all resources
func (p *LabelTransformer) setObjectMeta(o *fn.KubeObject) error {
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
		labelCountValue := labelCount[key]
		msg := fmt.Sprintf("set labels {%v} %v times", key, labelCountValue)
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

// parseLabelsFrom parses the `labelsFrom` input in the SetLabels functionConfig
// and returns pairs of label and resolved value.
func parseLabelsFrom(fnConf *fn.KubeObject, rl *fn.ResourceList) (map[string]string, error) {
	var labelSources []labelFrom

	found, err := fnConf.Get(&labelSources, "labelsFrom")
	if err != nil {
		return nil, err
	}
	if !found {
		// labelsFrom is an optional field
		return nil, nil
	}

	labels := map[string]string{}
	for _, labelSource := range labelSources {
		// iterate over the resources to extract values
		source := labelSource.Source
		matchedResources := rl.Items.Where(func(o *fn.KubeObject) bool {
			id := o.GetId()
			if source.Group != "" && (id.Group != source.Group) {
				return false
			}
			if source.Kind != "" && (id.Kind != source.Kind) {
				return false
			}
			if source.Name != "" && (id.Name != source.Name) {
				return false
			}
			// TODO: apply other selectors
			return true
		})
		if len(matchedResources) == 0 {
			// pick the first one or possible where package matches
			return nil, fmt.Errorf("couldn't find label source %s", source)
		}
		if len(matchedResources) > 1 {
			return nil, fmt.Errorf("multiple resources matched label source %s", source)
		}
		matchedResource := matchedResources[0]
		labelValue, found, err := matchedResource.NestedString(strings.Split(source.FieldPath, ".")...)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("fieldpath %q to extract labels does not exist in source %s", source.FieldPath, source)
		}
		labels[labelSource.Label] = labelValue
	}
	return labels, nil
}

// labelFrom represents the label and the source for the label value.
// It uses ApplyReplacement's source syntax so that users can avoid
// learning new syntax.
type labelFrom struct {
	// Label key
	Label string `json:"label" yaml:"label"`
	// Source of the label value
	Source types.SourceSelector `json:"source" yaml:"source"`
}
