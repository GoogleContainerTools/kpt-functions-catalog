package plugins

import (
	"sigs.k8s.io/kustomize/api/filters/annotations"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type AnnotationPlugin struct {
	Annotations map[string]string
	FieldSpecs  []types.FieldSpec `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

func (p *AnnotationPlugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Annotations = nil
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *AnnotationPlugin) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(annotations.Filter{
		Annotations: p.Annotations,
		FsSlice:     p.FieldSpecs,
	})
}
