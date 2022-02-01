package main

import (
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)
var _ kio.Filter = Filter{}

type Filter struct {
	ProjectID string
	FsSlice types.FsSlice `json:"projectID,omitempty" yaml:"projectID,omitempty"`
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

func (f Filter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	if err := node.PipeE(fsslice.Filter{
		FsSlice:  f.FsSlice,
		SetValue: f.updateProjectIDFn,
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (f Filter) updateProjectIDFn(node *yaml.RNode) error {
	return node.PipeE(updater{ProjectID: f.ProjectID,})
}

type updater struct {
	ProjectID string
}

func (u updater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	/*
	if err := yaml.ErrorIfInvalid(rn, yaml.ScalarNode); err != nil {
		return nil, err
	}
	 */
	// return fmt.Errorf("GET HERE %v %v", rn.GetKind())
	return rn.Pipe(yaml.FieldSetter{StringValue: u.ProjectID})
}


type ProjectIDTransformer struct {
	FieldSpecs []types.FieldSpec `json:"projectID,omitempty" yaml:"projectID,omitempty"`
}

func (p *ProjectIDTransformer) Transform(m resmap.ResMap, projectID string) error {
	return m.ApplyFilter(Filter{
		ProjectID: projectID,
		FsSlice:  p.FieldSpecs,
	})
}
