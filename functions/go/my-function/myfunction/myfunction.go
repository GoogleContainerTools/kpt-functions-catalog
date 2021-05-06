package myfunction

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/errors"
)

// FunctionConfig struct holds the input parameters for
// the function
type FunctionConfig struct {
	// Params holds the user provided inputs
	Params []Param `json:"params,omitempty" yaml:"params,omitempty"`
}

type Param struct {
	// Name is the name of the parameter
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Value is the input value for parameter
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// Filter performs the transformation/validation operation on all input yaml nodes
func (fc *FunctionConfig) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range nodes {
		err := accept(fc, nodes[i])
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return nodes, nil
}

// visitMapping parses mapping node and adds input comment to the mapping node key
func (fc *FunctionConfig) visitMapping(object *yaml.RNode, path string) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		/*
		// change the name and value of an annotation
		if path == "metadata.annotations.foo" {
			node.Key.YNode().Value = "abc"
			node.Value.YNode().Value = "def"
		}
		*/
		return nil
	})
}


// visitScalar parses input scalar node
func (fc *FunctionConfig) visitScalar(object *yaml.RNode, path string) error {
	/*
	// code snippet to increment the replicas count by 1
	if path == "metadata.replicas" {
		replicas, err := strconv.Atoi(object.YNode().Value)
		if err != nil {
			return err
		}
		object.YNode().Value = string(replicas+1)
	}
	*/


	/*
	// code snippet to impose constraint on setter value's length
	if object.YNode().LineComment == "# kpt-set: ${foo}" && len(object.YNode().Value) > 6 {
		return fmt.Errorf(`value of setter "foo" is longer than maximum allowed length "6"`)
	}
	*/
	return nil
}

// Decode decodes the input yaml node into Params struct
func Decode(rn *yaml.RNode, fcd *FunctionConfig) {
	for k, v := range rn.GetDataMap() {
		fcd.Params = append(fcd.Params, Param{Name: k, Value: v})
	}
}
