package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	//	"strings"

	//"io/ioutil"
	//"net/http"
	"os"
	"os/exec"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-labels/generated"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"

	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"

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

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"commonLabels,omitempty" yaml:"commonLabels,omitempty"`
}

type setLabelFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func (f *setLabelFunction) Config(rn *kyaml.RNode) error {
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
	if err := configureOpenAPI(); err != nil {
		return nil, err
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

func configureOpenAPI() error {
	openapi.SuppressBuiltInSchemaUse()

	cmd := exec.Command("sh", "-c", "ip route show | awk '/default/ {print $3}'")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("received error; error: %w, stdout: %s, stderr: %s", err.Error(), stdout.String(), stderr.String())
	}


	hostMachineIp := strings.TrimSpace(stdout.String())

	resp, err := http.Get("http://" + hostMachineIp + ":8080/OpenAPI")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	schema, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = openapi.AddSchema(schema); err != nil {
		return err
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

<<<<<<< HEAD:functions/go/set-labels/main.go
func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["commonlabels"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
=======
//nolint
func main() {
	fmt.Println("hello")
	resp, _ := http.Get("http://" +  "192.168.65.1:8080/OpenAPI")

	defer resp.Body.Close()

	schema, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(schema))
	return


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
>>>>>>> 86e8e51 (trying to curl host server):functions/go/set-label/main.go
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
