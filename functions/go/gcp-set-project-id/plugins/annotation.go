package plugins

import (
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-project-id/filter"
	"sigs.k8s.io/kustomize/api/resmap"
)

type AnnotationPlugin struct {
	Annotations map[string]string
}

func (p *AnnotationPlugin) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(filter.AnnoFilter{
		Annotations: p.Annotations,
	})
}
