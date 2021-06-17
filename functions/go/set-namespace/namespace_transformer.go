// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/namespace"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/yaml"
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
	}
	return nil
}
