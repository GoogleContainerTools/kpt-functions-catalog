package transformer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

// the similar struct esxits in resid.GVK, but there is no function to create an GVK struct without using kyaml
type FieldSpec struct {
	Group              string `json:"group,omitempty" yaml:"group,omitempty"`
	Version            string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind               string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Path               string `json:"path,omitempty" yaml:"path,omitempty"` // seperated by /
	CreateIfNotPresent bool   `json:"create,omitempty" yaml:"create,omitempty"`
	//TODO: should support user configurable field for labels
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
	// FieldSpecs storing default label fields
	FieldSpecs []FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// Results is used to track labels that have been applied
	Results fn.Results
}

// Config parse the functionConfig kubeObject to the fields in the LabelTransformer
func (p *LabelTransformer) Config(functionConfig *fn.KubeObject) error {
	// parse labels to NewLabels
	switch {
	case functionConfig.IsEmpty():
		return fmt.Errorf("failed to configure function: `functionConfig` must be either a `ConfigMap` or `SetLabels`")
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
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `%v` or `ConfigMap`",
			functionConfig.GetKind(), functionConfig.GetAPIVersion(), fnConfigKind)
	}
	return nil
}

// Transform updates the labels in the right path using configured logic
func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {

	for _, o := range objects {
		var path = "metadata/labels"
		err := updateLabels(o, path, p.NewLabels, true)
		if err != nil {
			return err
		}
		p.LogResult(o, path)
		// check all default path to update labels
		switch {
		case o.IsGVK("", "", ""):
			path = "metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "v1", "Service"):
			path = "spec/selector"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "v1", "ReplicationController"):
			path = "spec/selector"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "v1", "ReplicationController"):
			path = "spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "", "Deployment"):
			path = "spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "", "Deployment"):
			path = "spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("apps", "", "Deployment"):
			path = "spec/template/spec/affinity/podAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "Deployment"):
			path = "spec/template/spec/affinity/podAffinity/requiredDuringSchedulingIgnoredDuringExecution/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "Deployment"):
			path = "spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "Deployment"):
			path = "spec/template/spec/affinity/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "Deployment"):
			path = "spec/template/spec/topologySpreadConstraints/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("", "", "ReplicaSet"):
			path = "spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "", "ReplicaSet"):
			path = "spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "", "DaemonSet"):
			path = "spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("", "", "DaemonSet"):
			path = "spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/template/spec/affinity/podAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/template/spec/affinity/podAffinity/requiredDuringSchedulingIgnoredDuringExecution/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/template/spec/affinity/podAntiAffinity/requiredDuringSchedulingIgnoredDuringExecution/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/template/spec/topologySpreadConstraints/labelSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("apps", "", "StatefulSet"):
			path = "spec/volumeClaimTemplates[]/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("batch", "", "Job"):
			path = "spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("batch", "", "Job"):
			path = "spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("batch", "", "CronJob"):
			path = "spec/jobTemplate/spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("batch", "", "CronJob"):
			path = "spec/jobTemplate/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("batch", "", "CronJob"):
			path = "spec/jobTemplate/spec/template/metadata/labels"
			err = updateLabels(o, path, p.NewLabels, true)
		case o.IsGVK("policy", "", "PodDisruptionBudget"):
			path = "spec/selector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("networking.k8s.io", "", "NetworkPolicy"):
			path = "spec/podSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("networking.k8s.io", "", "NetworkPolicy"):
			path = "spec/ingress/from/podSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		case o.IsGVK("networking.k8s.io", "", "NetworkPolicy"):
			path = "spec/egress/to/podSelector/matchLabels"
			err = updateLabels(o, path, p.NewLabels, false)
		default:
			continue
		}

		if err != nil {
			return err
		}

		p.LogResult(o, path)
	}
	return nil
}

// Logs the result of each operation, can also modify into other logs user wants
func (p *LabelTransformer) LogResult(o *fn.KubeObject, path string) {
	res, _ := json.Marshal(p.NewLabels)
	newResult := fn.Result{
		Message:     "set labels: " + string(res),
		Severity:    "",
		ResourceRef: nil,
		Field: &fn.Field{
			Path:          path,
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

// the update process for each label, sort the keys to preserve sequence
func updateLabels(o *fn.KubeObject, fieldPath string, newLabels map[string]string, create bool) error {
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
