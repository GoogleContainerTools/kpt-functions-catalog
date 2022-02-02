package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-project-id/fieldspec"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ProjectAnnotationKey = "cnrm.cloud.google.com/project-id"

	// builtinAnnoKey is used by this package only.
	builtinAnnoKey = "commonannotations"
)

var (
	annoFieldSpecs = builtinpluginconsts.GetDefaultFieldSpecsAsMap()[builtinAnnoKey]
)

type ProjectIDTransformer struct {
	annotationPlugin  plugin
	projectFieldSpec types.FsSlice
	projectID string
}

func (p *ProjectIDTransformer) Config(fnConfigNode *yaml.RNode) error {
	// Get ProjectID
	data := fnConfigNode.GetDataMap()
	if data == nil {
		return fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	projectID, ok := data[projectIDKey]
	if !ok {
		return fmt.Errorf("missing `.data.%s` field in `ConfigMap` FunctionConfig", projectIDKey)
	}
	p.projectID = projectID

	// Enumerate project field specs
	var cfs CustomFieldSpec
	if err := yaml.Unmarshal([]byte(fieldspec.ProjectIDFieldSpecs), &cfs); err != nil {
		return err
	}
	p.projectFieldSpec = cfs.ProjectFieldSpecs

	// Set projectID in annotation plugin
	// TODO(yuwenma): Add option to exclude spec/template/metadata/annotations from builtin specfield.
	p.annotationPlugin.Annotations = map[string]string{ProjectAnnotationKey: projectID}
	var tc BuiltinAnnotationFilter
	if err := yaml.Unmarshal([]byte(annoFieldSpecs), &tc); err != nil {
		return err
	}
	p.annotationPlugin.AdditionalAnnotationFields = tc.FieldSpecs
	return nil
}

func (p *ProjectIDTransformer) Transform(m resmap.ResMap) error {
	if err := p.annotationPlugin.Transform(m); err != nil {
		return fmt.Errorf("plugin setAnnotation fail %v", err)
	}
	return m.ApplyFilter(CustomFieldSpecFilter{
		ProjectID: p.projectID,
		FsSlice:  p.projectFieldSpec,
	})
}

var _ kio.Filter = CustomFieldSpecFilter{}


type BuiltinAnnotationFilter struct {
	FieldSpecs types.FsSlice `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`
}

func (f CustomFieldSpecFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

type CustomFieldSpec struct {
	ProjectFieldSpecs []types.FieldSpec `json:"projectID,omitempty" yaml:"projectID,omitempty"`
}

type CustomFieldSpecFilter struct {
	ProjectID string
	FsSlice types.FsSlice `json:"projectID,omitempty" yaml:"projectID,omitempty"`
}

func (f CustomFieldSpecFilter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	if err := node.PipeE(fsslice.Filter{
		FsSlice:  f.FsSlice,
		SetValue: f.updateProjectIDFn,
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (f CustomFieldSpecFilter) updateProjectIDFn(node *yaml.RNode) error {
	return node.PipeE(updater{ProjectID: f.ProjectID,})
}

type updater struct {
	ProjectID string
}

func (u updater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	return rn.Pipe(yaml.FieldSetter{StringValue: u.ProjectID})
}

