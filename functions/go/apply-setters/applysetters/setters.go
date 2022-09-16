package applysetters

import (
	"fmt"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

const fnConfigKind = "ApplySetters"
const fnConfigApiVersion = "fn.kpt.dev/v1alpha1"

func ApplySetters(rl *fn.ResourceList) (bool, error) {
	r := Setters{}
	return r.Process(rl)
}

type Setters struct {
	// Setters holds the user provided values for all the setters
	Setters []Setter `yaml:"setters,omitempty" json:"setters,omitempty"`

	// Results are the results of applying setter values
	Results []*Result

	// filePath file path of resource
	filePath string
}

type Setter struct {
	// Name is the name of the setter
	Name string `yaml:"name" json:"name"`

	// Value is the input value for setter
	Value string `yaml:"value" json:"value"`
}

// Result holds result of search and replace operation
type Result struct {
	// FilePath is the file path of the matching field
	FilePath string

	// FieldPath is field path of the matching field
	FieldPath string

	// Value of the matching field
	Value string
}

// Config initializes Setters from a functionConfig fn.KubeObject
func (r *Setters) Config(functionConfig *fn.KubeObject) error {
	if functionConfig.IsEmpty() {
		return fmt.Errorf("FunctionConfig is missing. Expect `ApplySetters`")
	}
	if functionConfig.GetKind() != fnConfigKind || functionConfig.GetAPIVersion() != fnConfigApiVersion {
		return fmt.Errorf("received functionConfig of kind %s and apiVersion %s, "+
			"only functionConfig of kind %s and apiVersion %s is supported",
			functionConfig.GetKind(), functionConfig.GetAPIVersion(), fnConfigKind, fnConfigApiVersion)
	}
	r.Setters = []Setter{}
	if err := functionConfig.As(r); err != nil {
		return fmt.Errorf("unable to convert functionConfig to %s:\n%w",
			"setters", err)
	}
	return nil
}

// Process configures the setters and transforms them.
func (r *Setters) Process(rl *fn.ResourceList) (bool, error) {
	if err := r.Config(rl.FunctionConfig); err != nil {
		rl.LogResult(err)
		return false, nil
	}
	transformedItems, err := r.Transform(rl.Items)
	if err != nil {
		rl.LogResult(err)
		return false, nil
	}
	rl.Items = transformedItems
	return true, nil
}

// Transform runs the setters filter in order to apply the setters - this
// does the actual work.
func (r *Setters) Transform(items []*fn.KubeObject) ([]*fn.KubeObject, error) {
	var transformedItems []*fn.KubeObject
	var nodes []*yaml.RNode

	for _, obj := range items {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, objRN)
	}

	transformedNodes, err := r.Filter(nodes)
	if err != nil {
		return nil, err
	}
	for _, n := range transformedNodes {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return nil, err
		}
		transformedItems = append(transformedItems, obj)
	}
	return transformedItems, nil
}
