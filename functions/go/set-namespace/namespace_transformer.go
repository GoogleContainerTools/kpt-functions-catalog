// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/namespace"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

const (
	subjectsField          = "subjects"
	serviceAccountKind     = "ServiceAccount"
	roleBindingKind        = "RoleBinding"
	clusterRoleBindingKind = "ClusterRoleBinding"
)

// Change or set the namespace of non-cluster level resources.
type plugin struct {
	// Desired namespace.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalNamespaceFields instead.
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalNamespaceFields is used to specify additional fields to set namespace.
	AdditionalNamespaceFields []types.FieldSpec `json:"additionalNamespaceFields,omitempty" yaml:"additionalNamespaceFields,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Namespace = ""
	p.FieldSpecs = nil
	p.AdditionalNamespaceFields = nil
	if err = yaml.Unmarshal(c, p); err != nil {
		return fmt.Errorf("failed to unmarshal config %#v: %w", string(c), err)
	}
	if p.AdditionalNamespaceFields != nil && p.FieldSpecs != nil {
		return fmt.Errorf("`fieldSpecs` has been deprecated, please rename it to `additionalLabelFields`")
	}
	if p.AdditionalNamespaceFields == nil && p.FieldSpecs != nil {
		p.AdditionalNamespaceFields = p.FieldSpecs
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if len(p.Namespace) == 0 {
		return nil
	}
	sal := serviceAccountLookup{}
	sal.FromResMap(m)
	for _, r := range m.Resources() {
		if r.IsNilOrEmpty() {
			// Don't mutate empty objects?
			continue
		}
		err := filtersutil.ApplyToJSON(namespace.Filter{
			Namespace: p.Namespace,
			FsSlice:   p.AdditionalNamespaceFields,
		}, r)
		if err != nil {
			return err
		}
		if err = p.roleBindingHack(r, sal); err != nil {
			return err
		}
	}
	return nil
}

// roleBindingHack extends the behavior of the upstream kustomize filter
// This extension adds the following behavior:
// Given a ServiceAccount is present in the ResourceList, additionally set the
// namespace for that ServiceAccount where it is referenced in the "subjects"
// field of a RoleBinding or ClusterRoleBinding
func (p *plugin) roleBindingHack(r *resource.Resource, sal serviceAccountLookup) error {
	kind := r.GetKind()
	if kind != roleBindingKind && kind != clusterRoleBindingKind {
		return nil
	}
	subjects, err := r.Pipe(kyaml.Lookup(subjectsField))
	if err != nil || kyaml.IsMissingOrNull(subjects) {
		return err
	}

	err = subjects.VisitElements(func(o *kyaml.RNode) error {
		subjectKind := o.GetKind()
		if subjectKind != serviceAccountKind {
			return nil
		}
		var nameRN *kyaml.RNode
		nameRN, err = o.Pipe(kyaml.Lookup("name"))
		if err != nil || kyaml.IsMissingOrNull(nameRN) {
			return err
		}
		nameStr := kyaml.GetValue(nameRN)
		if !sal.HasServiceAccount(nameStr) {
			return nil
		}
		v := kyaml.NewScalarRNode(p.Namespace)
		return o.PipeE(
			kyaml.LookupCreate(kyaml.ScalarNode, "namespace"),
			kyaml.FieldSetter{Value: v},
		)
	})
	return err
}

// ServiceAccountLookup provides an API for tracking ServiceAccount resources
type serviceAccountLookup struct {
	serviceAccountMap map[string]bool
}

// FromResMap reads through a ResMap to populate ServiceAccountLookup
func (sal *serviceAccountLookup) FromResMap(m resmap.ResMap) {
	sal.serviceAccountMap = make(map[string]bool)
	for _, r := range m.Resources() {
		if r.GetKind() == serviceAccountKind {
			sal.serviceAccountMap[r.GetName()] = true
		}
	}
}

// HasServiceAccount returns whether a ServiceAccount with the provided name is present
func (sal *serviceAccountLookup) HasServiceAccount(name string) bool {
	return sal.serviceAccountMap[name]
}
