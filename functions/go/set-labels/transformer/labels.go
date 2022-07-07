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

// the similar struct esxits in resid.GVK, but there is no function to create an GVK struct without using kyaml
type FieldSpec struct {
	Identifier         GVK
	Path               FieldPath
	CreateIfNotPresent bool
	//TODO: should support user configurable field for labels
}

type FieldSpecs []FieldSpec

// generate common label paths
var Specs = FieldSpecs{
	FieldSpec{
		Identifier:         GVK{"", "v1", "Service"},
		Path:               FieldPath{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "v1", "ReplicationController"},
		Path:               FieldPath{"spec", "selector"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "v1", "ReplicationController"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "Deployment"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "Deployment"},
		Path:               FieldPath{"spec", "template", "spec", "topologySpreadConstraints", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "ReplicaSet"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "ReplicaSet"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "DaemonSet"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"", "", "DaemonSet"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "preferredDuringSchedulingIgnoredDuringExecution", "podAffinityTerm", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "affinity", "podAntiAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"apps", "", "StatefulSet"},
		Path:               FieldPath{"spec", "template", "spec", "topologySpreadConstraints", "labelSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "Job"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "Job"},
		Path:               FieldPath{"spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "CronJob"},
		Path:               FieldPath{"spec", "jobTemplate", "spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "CronJob"},
		Path:               FieldPath{"spec", "jobTemplate", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"batch", "", "CronJob"},
		Path:               FieldPath{"spec", "jobTemplate", "spec", "template", "metadata", "labels"},
		CreateIfNotPresent: true,
	},
	FieldSpec{
		Identifier:         GVK{"policy", "", "PodDisruptionBudget"},
		Path:               FieldPath{"spec", "selector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"networking.k8s.io", "", "NetworkPolicy"},
		Path:               FieldPath{"spec", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"networking.k8s.io", "", "NetworkPolicy"},
		Path:               FieldPath{"spec", "ingress", "from", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
	FieldSpec{
		Identifier:         GVK{"networking.k8s.io", "", "NetworkPolicy"},
		Path:               FieldPath{"spec", "egress", "to", "podSelector", "matchLabels"},
		CreateIfNotPresent: false,
	},
}

// perform the whole set labels operation according to given resourcelist
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

type LabelTransformer struct {
	// Desired labels
	NewLabels map[string]string
	// FieldSpecs storing default label fields
	FieldSpecs []FieldSpec
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

// set labels according to the generated common label
func (p *LabelTransformer) setLabelsInSpecs(o *fn.KubeObject) error {
	for _, spec := range Specs {
		if o.IsGVK(spec.Identifier.group, spec.Identifier.version, spec.Identifier.kind) {
			err := updateLabels(&o.SubObject, spec.Path, p.NewLabels, spec.CreateIfNotPresent)
			p.LogResult(o, spec.Path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Transform updates the labels in the right path using configured logic
func (p *LabelTransformer) Transform(objects fn.KubeObjects) error {
	for _, o := range objects {
		// this label need to set for all GKV
		defaultPath := FieldPath{"metadata", "labels"}
		err := updateLabels(&o.SubObject, defaultPath, p.NewLabels, true)
		p.LogResult(o, defaultPath)
		if err != nil {
			return err
		}
		// set other common labels according to specific GKV
		err = p.setLabelsInSpecs(o)
		if err != nil {
			return err
		}
		// handle special case with slice
		if o.IsGVK("apps", "", "StatefulSet") {
			for _, vctObj := range o.GetMap("spec").GetSlice("volumeClaimTemplates") {
				err = updateLabels(vctObj, defaultPath, p.NewLabels, true)
				p.LogResult(o, defaultPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Logs the result of each operation, can also modify into other logs user wants
func (p *LabelTransformer) LogResult(o *fn.KubeObject, path []string) {
	res, _ := json.Marshal(p.NewLabels)
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

// the update process for each label, sort the keys to preserve sequence
func updateLabels(o *fn.SubObject, labelPath FieldPath, newLabels map[string]string, create bool) error {
	keys := make([]string, 0)
	for k := range newLabels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := newLabels[key]
		newPath := append(labelPath, key)
		_, exist, err := o.NestedString(newPath...)
		if err != nil {
			return err
		}
		//TODO: should support user configurable field for labels
		if exist || create {
			if err = o.SetNestedString(val, newPath...); err != nil {
				return err
			}
		}
	}
	return nil
}
