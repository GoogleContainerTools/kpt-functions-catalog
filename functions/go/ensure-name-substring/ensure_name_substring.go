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
	fnConfigKind       = "EnsureNameSubstring"
)

type EnsureNameSubstring struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Substring is the desired name substring.
	Substring string `json:"substring" yaml:"substring"`
	// EditMode controls the desired action when the desired substring is not found in the name.
	// If not specified, prepend will be the default.
	EditMode EditMode `json:"editMode,omitempty" yaml:"editMode,omitempty"`

	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

type EditMode string

const (
	Prepend EditMode = "prepend"
	Append  EditMode = "append"
)

func (ens *EnsureNameSubstring) Defaults() {
	if ens.EditMode == "" {
		ens.EditMode = Prepend
	}
}

func (ens *EnsureNameSubstring) Validate() error {
	if len(ens.Substring) == 0 {
		return fmt.Errorf("substring must not be empty")
	}
	return nil
}

func (ens *EnsureNameSubstring) Transform(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		if shouldSkip(r.OrgId()) {
			continue
		}
		id := r.OrgId()
		// current default configuration contains
		// only one entry: "metadata/name" with no GVK
		for _, fs := range ens.FieldSpecs {
			if !id.IsSelected(&fs.Gvk) {
				continue
			}

			// Idempotency check: if the substring is already part of the name, we
			// don't need to do anything.
			hasSubstring, err := resourceContainsSubstring(r, ens.Substring, fs)
			if err != nil {
				return err
			}
			if hasSubstring {
				continue
			}

			if isNameChange(&fs) {
				// If we are changing "metadata/name", we tracks the original
				// name and the prefix or suffix being added.
				r.StorePreviousId()
				if ens.EditMode == Prepend {
					r.AddNamePrefix(ens.Substring)
				} else if ens.EditMode == Append {
					r.AddNameSuffix(ens.Substring)
				}
			}

			fltr := prefixsuffix.Filter{
				FieldSpec: fs,
			}
			if ens.EditMode == Prepend {
				fltr.Prefix = ens.Substring
			} else if ens.EditMode == Append {
				fltr.Suffix = ens.Substring
			}
			err = r.ApplyFilter(fltr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var _ yaml.Unmarshaler = &EnsureNameSubstring{}

func (ens *EnsureNameSubstring) UnmarshalYAML(value *yaml.Node) error {
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
		if err = configMapToEnsureNameSubstring(&cm, ens); err != nil {
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

func configMapToEnsureNameSubstring(cm *corev1.ConfigMap, ens *EnsureNameSubstring) error {
	if len(cm.Data) != 1 {
		return fmt.Errorf("only 1 entry is allowed in the ConfigMap, but got: %d", len(cm.Data))
	}
	for k, v := range cm.Data {
		switch k {
		case string(Prepend):
			ens.EditMode = Prepend
		case string(Append):
			ens.EditMode = Append
		default:
			return fmt.Errorf("unknown editMode: %v, only %v and %v are allowed", k, Prepend, Append)
		}
		ens.Substring = v
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

func resourceContainsSubstring(r *resource.Resource, substring string, fs types.FieldSpec) (bool, error) {
	if isNameChange(&fs) {
		// Idempotency check: if the substring is already part of the name, we
		// don't need to do anything.
		return strings.Contains(r.GetName(), substring), nil
	}

	rn, err := yaml.FromMap(r.Map())
	if err != nil {
		return false, fmt.Errorf("unable to check if the substring exsits in %v: %w", r.OrgId().String(), err)
	}
	pathElements := strings.Split(fs.Path, "/")
	val, err := rn.Pipe(yaml.Lookup(pathElements...))
	if err != nil {
		return false, fmt.Errorf("unable to lookup path %v in %v: %w", fs.Path, r.OrgId().String(), err)
	}
	valStr, err := val.String()
	if err != nil {
		return false, fmt.Errorf("unable to check if the substring exsits in %v: %w", r.OrgId().String(), err)
	}
	return strings.Contains(valStr, substring), nil
}
