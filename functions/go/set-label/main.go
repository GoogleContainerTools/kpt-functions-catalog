package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-label/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigVersion    = "v1alpha1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind       = "SetLabelConfig"

	FieldMeaningExtension = "x-kubernetes-field-meaning"
	Label                 = "label"
)

type setLabelFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func (f *setLabelFunction) Config(fnConfig interface{}) error {
	configMap, ok := fnConfig.(map[string]interface{})
	if !ok {
		return fmt.Errorf("function config %#v is not valid", fnConfig)
	}
	rn, err := kyaml.FromMap(configMap)
	if err != nil {
		return fmt.Errorf("failed to construct RNode from %#v: %w", configMap, err)
	}
	switch {
	case f.validGVK(rn, "v1", "ConfigMap"):
		f.Labels = rn.GetDataMap()
	case f.validGVK(rn, fnConfigAPIVersion, fnConfigKind):
		// input config is a CRD
		y, err := rn.String()
		if err != nil {
			return fmt.Errorf("cannot get YAML from RNode: %w", err)
		}
		err = yaml.Unmarshal([]byte(y), &f)
		if err != nil {
			return fmt.Errorf("failed to unmarshal config %#v: %w", y, err)
		}
	default:
		return fmt.Errorf("function config must be a ConfigMap or %s", fnConfigKind)
	}

	if len(f.Labels) == 0 {
		return fmt.Errorf("input label list cannot be empty")
	}
	return nil
}

func (f *setLabelFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	for _, r := range items {
		meta, _ := r.GetMeta()
		if meta.APIVersion == "kpt.dev/v1" && meta.Kind == "OpenAPI" {
			openapi.SuppressBuiltInSchemaUse()
			schema, err := r.Pipe(kyaml.Lookup("data"))
			if err != nil {
				return nil, fmt.Errorf("could not configure OpenAPI schema")
			}
			str, err := schema.String()
			if err != nil {
				return nil, err
			}
			json, err := yaml.YAMLToJSON([]byte(str))
			if err != nil {
				return nil, err
			}
			if err = openapi.AddSchema(json); err != nil {
				return nil, err
			}
		}
	}

	var result []*kyaml.RNode
	for _, r := range items {
		meta, _ := r.GetMeta()
		rs := openapi.SchemaForResourceType(meta.TypeMeta)
		if err := f.updateNode(r, rs); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (f *setLabelFunction) updateNode(node *kyaml.RNode, rs *openapi.ResourceSchema) error {
	if rs == nil || node == nil {
		return nil
	}

	for _, path := range getPathsToLabels(rs) {
		l, err := node.Pipe(kyaml.LookupCreate(kyaml.MappingNode, path...))
		if err != nil {
			return err
		}

		labels := make(map[string]interface{})
		_ = l.VisitFields(func(node *kyaml.MapNode) error {
			labels[kyaml.GetValue(node.Key)] = kyaml.GetValue(node.Value)
			return nil
		})
		for k, v := range f.Labels {
			labels[k] = v
		}

		n, err := kyaml.FromMap(labels)
		if err != nil {
			return err
		}

		l.SetYNode(n.YNode())
	}

	return nil
}

func getPathsToLabels(rs *openapi.ResourceSchema) [][]string {
	var result [][]string

	fieldMeaning, found := rs.Schema.Extensions[FieldMeaningExtension]
	if found && fieldMeaning == Label {
		result = append(result, []string{})
	}

	for field := range rs.Schema.Properties {
		for _, path := range getPathsToLabels(rs.Field(field)) {
			result = append(result, append([]string{field}, path...))
		}
	}
	return result
}


func (f *setLabelFunction) validGVK(rn *kyaml.RNode, apiVersion, kind string) bool {
	meta, err := rn.GetMeta()
	if err != nil {
		return false
	}
	if meta.APIVersion != apiVersion || meta.Kind != kind {
		return false
	}
	return true
}

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = map[string]interface{}{}

	cmd := framework.Command(resourceList, func() error {
		err := run(resourceList)
		if err != nil {
			resourceList.Result = &framework.Result{
				Name: "set-label",
				Items: []framework.Item{
					{
						Message:  err.Error(),
						Severity: framework.Error,
					},
				},
			}
			return resourceList.Result
		}
		return nil
	})

	cmd.Short = generated.SetLabelShort
	cmd.Long = generated.SetLabelLong
	cmd.Example = generated.SetLabelExamples
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(resourceList *framework.ResourceList) error {
	var fn setLabelFunction
	err := fn.Config(resourceList.FunctionConfig)
	if err != nil {
		return fmt.Errorf("failed to configure function: %w", err)
	}
	resourceList.Items, err = fn.Run(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to run function: %w", err)
	}
	return nil
}
