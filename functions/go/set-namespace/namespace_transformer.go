// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/namespace"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

const (
	subjectsField          = "subjects"
	serviceAccountKind     = "ServiceAccount"
	roleBindingKind        = "RoleBinding"
	clusterRoleBindingKind = "ClusterRoleBinding"
	dependsOnAnnotation    = "config.kubernetes.io/depends-on"
)

// Constants for namespaced resources
const (
	groupIdx     = 0
	namespaceIdx = 2
	kindIdx      = 3
	nameIdx      = 4
)

var (
	// Assumes alphanumeric characters, '-', or '.'
	// <group>/namespaces/<namespace>/<kind>/<name>
	namespacedResourcePattern = regexp.MustCompile(`\A([-.\w]*)/namespaces/([-.\w]*)/([-.\w]*)/([-.\w]*)\z`)
)

// Change or set the namespace of non-cluster level resources.
type plugin struct {
	// Desired namespace.
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalNamespaceFields instead.
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalNamespaceFields is used to specify additional fields to set namespace.
	AdditionalNamespaceFields []types.FieldSpec `json:"additionalNamespaceFields,omitempty" yaml:"additionalNamespaceFields,omitempty"`
	// inputResourceLookup is used internally to track input resources
	inputResourceLookup resourceLookup
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
	p.inputResourceLookup.FromResMap(m)
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
		if err = p.updateRoleBinding(r); err != nil {
			return err
		}
		if err = p.updateDependsOnAnnotation(r); err != nil {
			return err
		}
	}
	return nil
}

// set namespace if input matches <group>/namespaces/<namespace>/<kind>/<name>
// returns (string, bool) where the bool indicates if namespace was set
func (p *plugin) setDependsOnNamespace(dependsOn string) (string, bool) {
	if !namespacedResourcePattern.MatchString(dependsOn) {
		return dependsOn, false
	}
	segments := strings.Split(dependsOn, "/")
	rk := resourceKey{
		Group: segments[groupIdx],
		Kind:  segments[kindIdx],
		Name:  segments[nameIdx],
	}
	// Only update namespace if the referenced resource is included in input
	if !p.inputResourceLookup.HasResource(rk) {
		return dependsOn, false
	}
	segments[namespaceIdx] = p.Namespace
	dependsOn = strings.Join(segments, "/")
	return dependsOn, true
}

// updateDependsOnAnnotation updates the namespace for the depends-on annotation
// if the annotation is for a namespaced resource. The expected syntax for a
// namespaced resource is <group>/namespaces/<namespace>/<kind>/<name>
func (p *plugin) updateDependsOnAnnotation(r *resource.Resource) error {
	annotations := r.GetAnnotations()
	dependsOn, ok := annotations[dependsOnAnnotation]
	if !ok {
		return nil
	}
	dependsOn, ok = p.setDependsOnNamespace(dependsOn)
	if !ok {
		return nil
	}
	annotations[dependsOnAnnotation] = dependsOn
	if err := r.SetAnnotations(annotations); err != nil {
		return err
	}
	return nil
}

// updateRoleBinding extends the behavior of the upstream kustomize filter
// This extension adds the following behavior:
// Given a ServiceAccount is present in the ResourceList, additionally set the
// namespace for that ServiceAccount where it is referenced in the "subjects"
// field of a RoleBinding or ClusterRoleBinding
func (p *plugin) updateRoleBinding(r *resource.Resource) error {
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
		rk := resourceKey{}
		if err := rk.FromRNode(o); err != nil {
			return nil
		}
		if !p.inputResourceLookup.HasResource(rk) {
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

// resourceKey provides a key for looking up resources in resourceLookup
type resourceKey struct {
	Group string
	Kind  string
	Name  string
}

// FromRNode populates the fields of resourceKey from the provided RNode
func (rk *resourceKey) FromRNode(r *kyaml.RNode) error {
	gvk := resid.GvkFromNode(r)
	nameRN, err := r.Pipe(kyaml.Lookup("name"))
	if err != nil {
		return err
	}
	if kyaml.IsMissingOrNull(nameRN) {
		s, _ := r.String()
		return fmt.Errorf("RNode name missing or null from: %s", s)
	}
	rk.Kind = gvk.Kind
	rk.Group = gvk.Group
	rk.Name = kyaml.GetValue(nameRN)
	return nil
}

// resourceLookup provides an API for tracking resources
type resourceLookup struct {
	resourceMap map[resourceKey]bool
}

// FromResMap reads through a ResMap to populate ResourceLookup
func (rl *resourceLookup) FromResMap(m resmap.ResMap) {
	rl.resourceMap = make(map[resourceKey]bool)
	for _, r := range m.Resources() {
		gvk := r.GetGvk()
		rl.resourceMap[resourceKey{
			Group: gvk.Group,
			Kind:  gvk.Kind,
			Name:  r.GetName(),
		}] = true
	}
}

// HasResource returns whether a Resource with the provided resourceKey is present
func (rl *resourceLookup) HasResource(rk resourceKey) bool {
	return rl.resourceMap[rk]
}
