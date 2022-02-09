package filter

import (
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type annoMap map[string]string

type AnnoFilter struct {
	// Annotations is the set of annotations to apply to the inputs
	Annotations annoMap `yaml:"annotations,omitempty"`
}

var _ kio.Filter = AnnoFilter{}

func (f AnnoFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			existAnnos := node.GetAnnotations()
			for k, newVal := range f.Annotations {
				if _, ok := existAnnos[k]; ok {
					existAnnos[k] = newVal
				}
			}
			node.SetAnnotations(existAnnos)
			return node, nil
		})).Filter(nodes)
	return nodes, err
}
