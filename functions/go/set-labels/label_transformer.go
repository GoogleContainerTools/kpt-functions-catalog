// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/labels"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

// Add the given labels to the given field specifications.
type plugin struct {
	// Desired labels
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalLabelFields instead.
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalLabelFields is used to specify additional fields to add labels.
	AdditionalLabelFields []types.FieldSpec `json:"additionalLabelFields,omitempty" yaml:"additionalLabelFields,omitempty"`
	// Results is used to track labels that have been applied
	Results LabelResults
}

// LabelResults maps label paths to key/value pairs
type LabelResults map[LabelResultKey]LabelValues

// LabelResultKey is a unique representation for a label field path
type LabelResultKey struct {
	// FilePath is the file path of the resource
	FilePath string
	// FileIndex is the file index of the resource
	FileIndex string
	// FieldPath is field path of the labels
	FieldPath string
}

// LabelValues represents label key/value pairs
type LabelValues map[string]string

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Labels = nil
	p.FieldSpecs = nil
	p.AdditionalLabelFields = nil
	if err = yaml.Unmarshal(c, p); err != nil {
		return fmt.Errorf("failed to unmarshal config %#v: %w", string(c), err)
	}
	if p.AdditionalLabelFields != nil && p.FieldSpecs != nil {
		return fmt.Errorf("`fieldSpecs` has been deprecated, please rename it to `additionalLabelFields`")
	}
	if p.AdditionalLabelFields == nil && p.FieldSpecs != nil {
		p.AdditionalLabelFields = p.FieldSpecs
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if p.Results == nil {
		p.Results = make(LabelResults)
	}
	for _, r := range m.Resources() {
		filePath, fileIndex, err := kioutil.GetFileAnnotations(&r.RNode)
		if err != nil {
			return err
		}

		err = filtersutil.ApplyToJSON(labels.Filter{
			Labels:  p.Labels,
			FsSlice: p.AdditionalLabelFields,
			SetEntryCallback: func(key, value, tag string, node *kyaml.RNode) {
				resultKey := LabelResultKey{
					FieldPath: strings.Join(node.FieldPath(), "."),
					FilePath:  filePath,
					FileIndex: fileIndex,
				}
				result, ok := p.Results[resultKey]
				if ok {
					result[key] = value
				} else {
					p.Results[resultKey] = LabelValues{key: value}
				}
			},
		}, r)
		if err != nil {
			return err
		}
	}
	return nil
}
