package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-project-id/consts"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-project-id/plugins"
	"sigs.k8s.io/kustomize/api/resmap"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ProjectAnnotationKey = "cnrm.cloud.google.com/project-id"
)

type ProjectIDTransformer struct {
	annotation plugins.AnnotationPlugin
	custom     plugins.CustomFieldSpecPlugin
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

	// Set transformer to update custom field spec paths
	if err := p.custom.Config([]byte(consts.ProjectFieldSpecs)); err != nil {
		return err
	}
	p.custom.ProjectID = projectID

	// Set transformer to update annotation paths (metadata and inner resources' metadata).
	p.annotation.Annotations = map[string]string{ProjectAnnotationKey: projectID}
	return nil
}

func (p *ProjectIDTransformer) Transform(m resmap.ResMap) error {
	if err := p.annotation.Transform(m); err != nil {
		return fmt.Errorf("set projectID to annotations fail %v", err)
	}
	if err := p.custom.Transform(m); err != nil {
		return fmt.Errorf("set projectID to custom fieldSpec fail %v", err)
	}
	return nil
}
