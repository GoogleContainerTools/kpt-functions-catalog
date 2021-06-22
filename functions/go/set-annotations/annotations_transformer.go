// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/filters/annotations"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Add the given annotations to the given field specifications.
type plugin struct {
	// Desired annotations
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalAnnotationFields instead.
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalAnnotationFields is used to specify additional fields to add annotations.
	AdditionalAnnotationFields []types.FieldSpec `json:"additionalAnnotationFields,omitempty" yaml:"additionalAnnotationFields,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Annotations = nil
	p.FieldSpecs = nil
	p.AdditionalAnnotationFields = nil
	if err = yaml.Unmarshal(c, p); err != nil {
		return fmt.Errorf("failed to unmarshal config %#v: %w", string(c), err)
	}
	if p.AdditionalAnnotationFields != nil && p.FieldSpecs != nil {
		return fmt.Errorf("`fieldSpecs` has been deprecated, please rename it to `additionalAnnotationFields`")
	}
	if p.AdditionalAnnotationFields == nil && p.FieldSpecs != nil {
		p.AdditionalAnnotationFields = p.FieldSpecs
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if len(p.Annotations) == 0 {
		return nil
	}
	for _, r := range m.Resources() {
		err := r.ApplyFilter(annotations.Filter{
			Annotations: p.Annotations,
			FsSlice:     p.AdditionalAnnotationFields,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
