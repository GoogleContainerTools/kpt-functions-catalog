// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
// Copied and modified from
// https://github.com/kubernetes-sigs/kustomize/blob/3265f64cd5ea76a8b64877b193576e2d120001db/api/builtins/PrefixSuffixTransformer.go

//go:generate pluginator
package main

import (
	"strings"

	"sigs.k8s.io/kustomize/api/filters/prefixsuffix"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Add the given prefix and suffix to the field.
type plugin struct {
	Prefix     string        `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	FieldSpecs types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

// A Gvk skip list for prefix/suffix modification.
// hard coded for now - eventually should be part of config.
var prefixSuffixFieldSpecsToSkip = types.FsSlice{
	{Gvk: resid.Gvk{Kind: "CustomResourceDefinition"}},
	{Gvk: resid.Gvk{Group: "apiregistration.k8s.io", Kind: "APIService"}},
	{Gvk: resid.Gvk{Kind: "Namespace"}},
}

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Prefix = ""
	p.FieldSpecs = nil
	err = yaml.Unmarshal(c, p)
	if err != nil {
		return
	}
	return
}

func (p *plugin) Transform(m resmap.ResMap) error {
	// Even if both the Prefix and Suffix are empty we want
	// to proceed with the transformation. This allows to add contextual
	// information to the resources (AddNamePrefix and AddNameSuffix).
	for _, r := range m.Resources() {
		// TODO: move this test into the filter (i.e. make a better filter)
		if p.shouldSkip(r.OrgId()) {
			continue
		}
		id := r.OrgId()
		// current default configuration contains
		// only one entry: "metadata/name" with no GVK
		for _, fs := range p.FieldSpecs {
			// TODO: this is redundant to filter (but needed for now)
			if !id.IsSelected(&fs.Gvk) {
				continue
			}
			// check idempotent
			if strings.HasPrefix(r.GetName(), p.Prefix) {
				continue
			}

			// TODO: move this test into the filter.
			if smellsLikeANameChange(&fs) {
				// "metadata/name" is the only field.
				// this will add a prefix and a suffix
				// to the resource even if those are
				// empty

				r.AddNamePrefix(p.Prefix)
				if p.Prefix != "" {
					r.StorePreviousId()
				}
			}
			err := r.ApplyFilter(prefixsuffix.Filter{
				Prefix:    p.Prefix,
				FieldSpec: fs,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func smellsLikeANameChange(fs *types.FieldSpec) bool {
	return fs.Path == "metadata/name"
}

func (p *plugin) shouldSkip(id resid.ResId) bool {
	for _, path := range prefixSuffixFieldSpecsToSkip {
		if id.IsSelected(&path.Gvk) {
			return true
		}
	}
	return false
}
