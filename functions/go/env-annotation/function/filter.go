package function

import (
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Config struct {
	Data map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
}

type Filter struct {
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

func NewFilter(cfg *Config) (kio.Filter, error) {
	f := Filter{Annotations: map[string]string{}}
	for ai, av := range cfg.Data {
		if av == "" {
			f.Annotations[ai] = os.Getenv(ai)
		} else {
			f.Annotations[ai] = av
		}
	}
	return &f, nil
}

func (f *Filter) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, r := range items {
		for ai, av := range f.Annotations {
			err := r.PipeE(yaml.SetAnnotation(ai, av))
			if err != nil {
				return nil, fmt.Errorf("Couldn't set annotation %s=%s: %v", ai, av, err)
			}
		}
	}
	return items, nil
}
