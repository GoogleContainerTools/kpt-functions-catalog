// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
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
	// Results are the results of applying setter values
	Results []*Result
	// filePath file path of resource
	filePath string
}

// Result holds result of search and replace operation
type Result struct {
	// FilePath is the file path of the matching field
	FilePath string
	// FieldPath is field path of the matching field
	FieldPath string
	// Value of the matching field
	Value string
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

// setEntry tracks mutated fields in Results and calls filtersutil.SetEntry
func (p *plugin) setEntry(key, value, tag string) filtersutil.SetFn {
	baseSetEntry := filtersutil.SetEntry(key, value, tag)
	return func(node *kyaml.RNode) error {
		p.Results = append(p.Results, &Result{
			Value: value,
			FieldPath: strings.Join(append(node.FieldPath(), key), "."),
			FilePath: p.filePath,
		})
		return baseSetEntry(node)
	}
}

// Filter implements the kio.Filter interface to update annotations using setEntry
func (p *plugin) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	keys := kyaml.SortedMapKeys(p.Annotations)
	_, err := kio.FilterAll(kyaml.FilterFunc(
		func(node *kyaml.RNode) (*kyaml.RNode, error) {
			for _, k := range keys {
				if err := node.PipeE(fsslice.Filter{
					FsSlice: p.AdditionalAnnotationFields,
					SetValue: p.setEntry(
						k, p.Annotations[k], kyaml.NodeTagString),
					CreateKind: kyaml.MappingNode, // Annotations are MappingNodes.
					CreateTag:  kyaml.NodeTagMap,
				}); err != nil {
					return nil, err
				}
			}
			return node, nil
		})).Filter(nodes)
	return nodes, err
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
		p.filePath = filePath
		err = r.ApplyFilter(p)
		if err != nil {
			return err
		}
	}
	return nil
}
