package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-labels/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
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

//nolint
func main() {
	resourceList := &framework.ResourceList{}
	resourceList.FunctionConfig = &kyaml.RNode{}
	asp := SetLabelsProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.SetLabelsShort
	cmd.Long = generated.SetLabelsLong
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type SetLabelsProcessor struct{}

func (slp *SetLabelsProcessor) Process(resourceList *framework.ResourceList) error {
	err := run(resourceList)
	if err != nil {
		resourceList.Result = &framework.Result{
			Name: "set-labels",
			Items: []framework.ResultItem{
				{
					Message:  err.Error(),
					Severity: framework.Error,
				},
			},
		}
		return resourceList.Result
	}
	return nil
}

type setLabelFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func (f *setLabelFunction) Config(rn *kyaml.RNode) error {
	if err := configureOpenAPI(); err != nil {
		return err
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

func configureOpenAPI() error {
	openapi.SuppressBuiltInSchemaUse()

	req, err := http.NewRequest("GET", "http://host.docker.internal:8080/openapi", nil)
	if err != nil {
		return err
	}
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("could not make curl request: %w", err)
	}

	if resp != nil && resp.Body != nil {
		if resp.StatusCode != 200 {
			return fmt.Errorf("response from local server: %s", resp.Status)
		}

		schema, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not get response body: %w", err)
		}
		s := string(schema)
		s = strings.ReplaceAll(s, "(MISSING)", "")
		s = strings.ReplaceAll(s, "!", " ")

		if err = openapi.AddSchema([]byte(s)); err != nil {
			return fmt.Errorf("could not parse schema: %w %s", err, s)
		}
		_ = resp.Body.Close()
	}
	return nil
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