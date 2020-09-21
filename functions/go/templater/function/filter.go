package function

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Config struct {
	// Data contains map with object parameters to render
	// it's compatible with ConfigMap object, but isn't limited to strings
	// there is a possibility to use complex objects
	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`
}

type Filter struct {
	Template string                 `json:"template" yaml:"template"`
	Data     map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`
}

func NewFilter(cfg *Config) (kio.Filter, error) {
	val, ok := cfg.Data["template"]
	if !ok {
		return nil, fmt.Errorf("config doesn't have data.template field: %v", cfg)
	}

	template, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("data.template must be string")
	}

	f := Filter{Template: template, Data: cfg.Data}
	delete(f.Data, "template")
	return &f, nil
}

func (f *Filter) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var out bytes.Buffer

	funcMap := sprig.TxtFuncMap()
	funcMap["toYaml"] = toYaml
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(f.Template)
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(&out, f.Data)
	if err != nil {
		return nil, fmt.Errorf("template returned error: %v", err)
	}

	// Convert string to Rnodes
	b := kio.PackageBuffer{}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: bytes.NewBufferString(out.String())}},
		Outputs: []kio.Writer{&b},
	}
	err = p.Execute()
	if err != nil {
		return nil, err
	}
	return append(items, b.Nodes...), nil
}

// Render input yaml as output yaml
func toYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}
