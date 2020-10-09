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
	Entrypoint string                 `json:"entrypoint" yaml:"entrypoint"`
	Data       map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`
}

func NewFilter(cfg *Config) (kio.Filter, error) {
	val, ok := cfg.Data["entrypoint"]
	if !ok {
		return nil, fmt.Errorf("config doesn't have data.entrypoint field: %v", cfg)
	}

	entrypoint, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("data.entrypoint must be string")
	}

	f := Filter{Entrypoint: entrypoint, Data: cfg.Data}
	delete(f.Data, "entrypoint")
	return &f, nil
}

func (f *Filter) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	var out bytes.Buffer

	funcMap := sprig.TxtFuncMap()
	funcMap = FuncMapMerge(funcMap, FuncMap())
	funcMap["toYaml"] = toYaml
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(f.Entrypoint)
	if err != nil {
		return nil, err
	}

	tmplRoot := map[string]interface{}{
		"Items": items,
		"Data":  f.Data,
	}

	err = tmpl.Execute(&out, tmplRoot)
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
		return nil, fmt.Errorf("can't unmarshal:\n%s\n, %v", out.String(), err)
	}
	items, err = getRNodes(tmplRoot["Items"])
	if err != nil {
		return nil, fmt.Errorf("Can't convert Items back: %v", err)
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

func getRNodes(rnodesarr interface{}) ([]*yaml.RNode, error) {
	rnodes, ok := rnodesarr.([]*yaml.RNode)
	if ok {
		return rnodes, nil
	}

	rnodesx, ok := rnodesarr.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type %T - wanted []", rnodesarr)
	}

	rns := []*yaml.RNode{}
	for i, r := range rnodesx {
		rn, ok := r.(*yaml.RNode)
		if !ok {
			return nil, fmt.Errorf("has got element %d with unexpected type %T", i, r)
		}
		rns = append(rns, rn)
	}
	return rns, nil
}
