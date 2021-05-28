package createsetters

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &CreateSetters{}

// CreateSetters applies the setter values to the resource fields which are tagged
// by the setter reference comments
type CreateSetters struct {
	// Setters holds the user provided values for all the setters
	Setters []Setter

	// Setters holds the user provided values for array setters
	ArraySetters []ArraySetter

	// Results are the results of applying setter values
	Results []*Result

	// filePath file path of resource
	filePath string
}

type Setter struct {
	// Name is the name of the setter
	Name string

	// Value is the input value for setter
	Value string
}

// ArraySetter hi
type ArraySetter struct {

	Name string

	ValueSet map[string]bool
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

// Filter implements Set cs a yaml.Filter
func (cs *CreateSetters) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	if len(cs.Setters) == 0 {
		return nodes, fmt.Errorf("input setters list cannot be empty")
	}
	for i := range nodes {
		filePath, _, err := kioutil.GetFileAnnotations(nodes[i])
		if err != nil {
			return nodes, err
		}
		cs.filePath = filePath
		err = accept(cs, nodes[i])
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return nodes, nil
}


func (cs *CreateSetters) visitMapping(object *yaml.RNode, path string) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		if node == nil || node.Key.IsNil() || node.Value.IsNil() {
			// don't do IsNilOrEmpty check cs empty sequences are allowed
			return nil
		}

		
		if node.Value.YNode().Kind != yaml.SequenceNode {
			return nil
		}

		changeNode := node.Key

		if node.Value.YNode().Style == yaml.FlowStyle {
			changeNode = node.Value
		}

		elements, err := node.Value.Elements()
		if err != nil {
			return errors.Wrap(err)
		}
		var str []string
		for _, ar := range elements {
			str = append(str, ar.YNode().Value)
		}

		for _, arraySetters := range cs.ArraySetters {
			
			if checkEqual(str, arraySetters.ValueSet) {
				changeNode.YNode().LineComment = fmt.Sprintf("# kpt-set: ${%s}", arraySetters.Name)
				return nil
			}
		}

		return nil
	})
}


func (cs *CreateSetters) visitScalar(object *yaml.RNode, path string) error {
	if object.YNode().Kind != yaml.ScalarNode {
		// return if it is not a scalar node
		return nil
	}

	if path[len(path)-1] != ']' {
		contains := false

		linecomment := object.YNode().Value

		for _, setter := range cs.Setters {

			if strings.Contains(object.YNode().Value, setter.Value) {
				contains = true
				linecomment = strings.ReplaceAll(
					linecomment,
					setter.Value,
					fmt.Sprintf("${%s}", setter.Name),
				)
			}
		}
		if contains {
			object.YNode().LineComment = fmt.Sprintf("kpt-set: %s", linecomment)
		}
	}
	return nil
}

// Decode decodes the input yaml node into Set struct
func Decode(rn *yaml.RNode, fcd *CreateSetters) {
	for k, v := range rn.GetDataMap() {
		if isArraySetter(v) {
			fcd.ArraySetters = append(fcd.ArraySetters, ArraySetter{Name: k, ValueSet: getArraySetter(v)})
		} else {
			fcd.Setters = append(fcd.Setters, Setter{Name: k, Value: v})
		}
	}
}

func checkEqual(a []string, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}

	for _, aEle := range a{
		if !b[aEle] {
			return false
		} 
	}
	return true
}

func getArraySetter(input string) map[string]bool {
	output := make(map[string]bool)

	sv, err := yaml.Parse(input)
	if err != nil {
		return output
	}

	ele, err := sv.Elements()
	if err != nil {
		return output
	}
	
	for _, ar := range ele {
		output[ar.YNode().Value] = true
	}
	
	return output
}

func isArraySetter(value string) bool {
	if strings.Contains(value, "-") || strings.Contains(value, "[") {
		return true
	}
	return false
}
