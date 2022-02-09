package plugins

import (
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-project-id/custom"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type CustomFieldSpecPlugin struct {
	ProjectID string
	FsSlice   []types.FieldSpec `json:"projectFieldSpec,omitempty" yaml:"projectFieldSpec,omitempty"`
}

func (f *CustomFieldSpecPlugin) Config(c []byte) error {
	f.FsSlice = nil
	return yaml.Unmarshal(c, f)
}

func (f *CustomFieldSpecPlugin) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(custom.Filter{
		ProjectID: f.ProjectID,
		FsSlice:   f.FsSlice,
	})
}
