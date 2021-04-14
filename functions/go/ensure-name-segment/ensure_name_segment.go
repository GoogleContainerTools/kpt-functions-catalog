package main

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/filters/prefixsuffix"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigVersion    = "v1alpha1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "EnsureNameSegment"
)

type EnsureNameSegment struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Segment is the desired name segment.
	Segment string `json:"segment" yaml:"segment"`
	// ActionOnNotFound controls the desired action when the desired segment is not found in the name.
	// If not specified, prepend will be the default.
	ActionOnNotFound ActionMode `json:"actionOnNotFound,omitempty" yaml:"actionOnNotFound,omitempty"`

	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

type ActionMode string

const (
	Prepend ActionMode = "prepend"
	Append  ActionMode = "append"
)

func (ens *EnsureNameSegment) Defaults() {
	if ens.ActionOnNotFound == "" {
		ens.ActionOnNotFound = Prepend
	}
}

func (ens *EnsureNameSegment) Validate() error {
	if len(ens.Segment) == 0 {
		return fmt.Errorf("segment must not be empty")
	}
	return nil
}

func (ens *EnsureNameSegment) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		if shouldSkip(r.OrgId()) {
			continue
		}
		id := r.OrgId()
		// current default configuration contains
		// only one entry: "metadata/name" with no GVK
		for _, fs := range ens.FieldSpecs {
			// TODO: this is redundant to filter (but needed for now)
			if !id.IsSelected(&fs.Gvk) {
				continue
			}

			// Idempotency check: if the segment is already part of the name, we
			// don't need to do anything.
			contain, err := resourceContainsSegment(r, ens.Segment, fs)
			if err != nil {
				return err
			}
			if contain {
				continue
			}

			if isNameChange(&fs) {
				// If we are changing "metadata/name", we tracks the original
				// name and the prefix or suffix being added.
				r.StorePreviousId()
				if ens.ActionOnNotFound == Prepend {
					r.AddNamePrefix(ens.Segment)
				} else if ens.ActionOnNotFound == Append {
					r.AddNameSuffix(ens.Segment)
				}
			}

			fltr := prefixsuffix.Filter{
				FieldSpec: fs,
			}
			if ens.ActionOnNotFound == Prepend {
				fltr.Prefix = ens.Segment
			} else if ens.ActionOnNotFound == Append {
				fltr.Suffix = ens.Segment
			}
			err = r.ApplyFilter(fltr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var _ yaml.Unmarshaler = &EnsureNameSegment{}

func (ens *EnsureNameSegment) UnmarshalYAML(value *yaml.Node) error {
	rn := yaml.NewRNode(value)
	meta, err := rn.GetValidatedMetadata()
	if err != nil {
		return err
	}
	s, err := rn.String()
	if err != nil {
		return err
	}

	switch {
	case meta.APIVersion == "v1" && meta.Kind == "ConfigMap":
		cm := corev1.ConfigMap{}
		if err = k8syaml.Unmarshal([]byte(s), &cm); err != nil {
			return err
		}
		if err = configMapToEnsureNameSegment(&cm, ens); err != nil {
			return err
		}
	case meta.APIVersion == fnConfigAPIVersion && meta.Kind == fnConfigKind:
		if err = k8syaml.Unmarshal([]byte(s), ens); err != nil {
			return err
		}
	default:
		return fmt.Errorf("function config must be either %v or %v, but we got: %v",
			schema.FromAPIVersionAndKind("v1", "ConfigMap").String(),
			schema.FromAPIVersionAndKind(fnConfigAPIVersion, fnConfigKind).String(),
			schema.FromAPIVersionAndKind(meta.APIVersion, meta.Kind).String())
	}
	return nil
}

func configMapToEnsureNameSegment(cm *corev1.ConfigMap, ens *EnsureNameSegment) error {
	if len(cm.Data) != 1 {
		return fmt.Errorf("only 1 entry is allowed in the ConfigMap, but got: %d", len(cm.Data))
	}
	for k, v := range cm.Data {
		switch k {
		case string(Prepend):
			ens.ActionOnNotFound = Prepend
		case string(Append):
			ens.ActionOnNotFound = Append
		default:
			return fmt.Errorf("actionOnNotFound")
		}
		ens.Segment = v
	}
	return nil
}

// A Gvk skip list for prefix/suffix modification.
// hard coded for now - eventually should be part of config.
var prefixSuffixFieldSpecsToSkip = types.FsSlice{
	{Gvk: resid.Gvk{Kind: "CustomResourceDefinition"}},
	{Gvk: resid.Gvk{Group: "apiregistration.k8s.io", Kind: "APIService"}},
	{Gvk: resid.Gvk{Kind: "Namespace"}},
}

func shouldSkip(id resid.ResId) bool {
	for _, path := range prefixSuffixFieldSpecsToSkip {
		if id.IsSelected(&path.Gvk) {
			return true
		}
	}
	return false
}

func isNameChange(fs *types.FieldSpec) bool {
	return fs.Path == "metadata/name"
}

func resourceContainsSegment(r *resource.Resource, segment string, fs types.FieldSpec) (bool, error) {
	if isNameChange(&fs) {
		// Idempotency check: if the segment is already part of the name, we
		// don't need to do anything.
		return strings.Contains(r.GetName(), segment), nil
	}

	rn, err := yaml.FromMap(r.Map())
	if err != nil {
		return false, err
	}
	pathElements := strings.Split(fs.Path, "/")
	val, err := rn.Pipe(yaml.Lookup(pathElements...))
	if err != nil {
		return false, err
	}
	valStr, err := val.String()
	if err != nil {
		return false, err
	}
	return strings.Contains(valStr, segment), nil
}
