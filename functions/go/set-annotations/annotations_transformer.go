// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/annotations"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
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
	// Results are the results of applying annotations
	Results []*Result
}

// Result holds result of set annotation operation
type Result struct {
	// FilePath is the file path of the annotation
	FilePath string
	// FieldPath is field path of the annotation
	FieldPath string
	// Value of the annotation
	Value string
}

func (r *Result) String() string {
	return fmt.Sprintf("FilePath: '%s', FieldPath: '%s', Value: '%s'\n",
		r.FilePath, r.FieldPath, r.Value)
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
		filePath, _, err := kioutil.GetFileAnnotations(&r.RNode)
		if err != nil {
			return err
		}
		err = r.ApplyFilter(annotations.Filter{
			FsSlice:     p.AdditionalAnnotationFields,
			Annotations: p.Annotations,
			SetEntryCallback: func(key, value, tag string, node *kyaml.RNode) {
				p.Results = append(p.Results, &Result{
					Value:     value,
					FieldPath: strings.Join(append(node.FieldPath(), key), "."),
					FilePath:  filePath,
				})
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
