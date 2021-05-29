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

// CreateSetters creates a comment for the resource fields which 
// contains the same value as setter reference value 
type CreateSetters struct {
	// Setters holds the user provided values for simple map setters
	Setters []Setter

	// Setters holds the user provided values for array setters
	ArraySetters []ArraySetter

	// Results are the results of applying setter values
	Results []*Result

	// filePath file path of resource
	filePath string
}

// Setter stores name and value of the map setter
type Setter struct {
	// Name is the name of the setter
	Name string

	// Value is the input value for setter
	Value string
}

// ArraySetter stores name and values of the array setter
type ArraySetter struct {
	// Name is the name of the setter
	Name string

	// ValueSet is the set of the values for setter
	ValueSet map[string]bool
}

// Result holds result of search and replace operation
type Result struct {
	// FilePath is the file path of the matching value
	FilePath string

	// FieldPath is field path of the matching value
	FieldPath string

	// Comment of the matching value
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

/**
visitMapping takes the mapping node and performs following steps,
checks if it is a sequence node 
checks if all the values in the node are present to any of the ArraySetters
adds the linecomment if they are equal 

e.g. for input of Mapping node

environments: 
  - dev
  - stage

For input ApplySetters [Name: env, ValueSet: [dev, stage]], yaml node is transformed to

environments: # kpt-set: ${env}
  - dev
  - stage

e.g. for input of Mapping node with FlowStyle

env: [foo, bar]

For input ApplySetters [Name: env, ValueSet: [foo, bar]], yaml node is transformed to

env: [foo, bar] # kpt-set: ${env}
*/

func (cs *CreateSetters) visitMapping(object *yaml.RNode, path string) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		if node == nil || node.Key.IsNil() || node.Value.IsNil() {
			// don't do IsNilOrEmpty check cs empty sequences are allowed
			return nil
		}
		
		// the aim of this method is to create-setter for sequence nodes
		if node.Value.YNode().Kind != yaml.SequenceNode {
			// return if it is not a sequence node
			return nil
		}

		// checks if the kind is flowstyle and adds comment to its value node 
		// else it adds the comment to the key node 
		changeNodeComment := node.Key

		if node.Value.YNode().Style == yaml.FlowStyle {
			changeNodeComment = node.Value
		}

		// add the key to the field path
		fieldPath := strings.TrimPrefix(fmt.Sprintf("%s.%s", path, node.Key.YNode().Value), ".")

		elements, err := node.Value.Elements()
		if err != nil {
			return errors.Wrap(err)
		}
		// extracts the values in sequence node to an array 
		var nodeValues []string
		for _, values := range elements {
			nodeValues = append(nodeValues, values.YNode().Value)
		}

		for _, arraySetters := range cs.ArraySetters {
			// checks if all the values in node are present in array setter
			if checkEqual(nodeValues, arraySetters.ValueSet) {
				changeNodeComment.YNode().LineComment = fmt.Sprintf("kpt-set: ${%s}", arraySetters.Name)
				return nil
			}
		}

		cs.Results = append(cs.Results, &Result{
			FilePath:  cs.filePath,
			FieldPath: fieldPath,
			Value: changeNodeComment.YNode().LineComment,
		})
		return nil
	})
}

/**
visitScalar accepts the input scalar node and performs following steps,
checks if it is a scalar node 
checks if the path ends with ']' (which is the value of a sequence node)
adds the linecomment if it's value matches with any of the setter 

e.g.for input of scalar node 'nginx:1.7.1' in the yaml node

apiVersion: v1
...
  image: nginx:1.7.1

and for input CreateSetters [[name: image, value: nginx], [name: tag, value: 1.7.1]]
The yaml node is transformed to

apiVersion: v1
...
  image: nginx:1.7.1 # kpt-set: ${image}:${tag}

*/

func (cs *CreateSetters) visitScalar(object *yaml.RNode, path string) error {
	if object.YNode().Kind != yaml.ScalarNode {
		// return if it is not a scalar node
		return nil
	}

	// checks if the path ends with "]"
	if path[len(path)-1] == ']' {
		return nil
	}

	// its a flag to indicate if value matches with any of the setter 
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

	// sets the linecomment
	if contains {
		object.YNode().LineComment = fmt.Sprintf("kpt-set: %s", linecomment)
	}

	cs.Results = append(cs.Results, &Result{
		FilePath:  cs.filePath,
		FieldPath: strings.TrimPrefix(path, "."),
		Value:  object.YNode().LineComment,
	})
	
	return nil
}

// Decode decodes the input yaml node into Set struct
func Decode(rn *yaml.RNode, fcd *CreateSetters) {
	for k, v := range rn.GetDataMap() {
		// add the setter to ArraySetters if value is array 
		// else add to the Setters
		if isArraySetter(v) {
			fcd.ArraySetters = append(fcd.ArraySetters, ArraySetter{Name: k, ValueSet: getArraySetter(v)})
		} else {
			fcd.Setters = append(fcd.Setters, Setter{Name: k, Value: v})
		}
	}
}

// checkEqual checks if all the values in node are present in array setter
func checkEqual(nodeValues []string, arraySetters map[string]bool) bool {
	if len(nodeValues) != len(arraySetters) {
		return false
	}

	for _, value := range nodeValues{
		if !arraySetters[value] {
			return false
		} 
	}
	return true
}

// parses the input and returns array setters
func getArraySetter(input string) map[string]bool {
	output := make(map[string]bool)

	parsedInput, err := yaml.Parse(input)
	if err != nil {
		return output
	}

	elements, err := parsedInput.Elements()
	if err != nil {
		return output
	}
	
	for _, as := range elements {
		output[as.YNode().Value] = true
	}
	
	return output
}

// isArraySetter checks if it is a array setter 
func isArraySetter(value string) bool {
	if strings.Contains(value, "- ") || strings.Contains(value, "[") {
		return true
	}
	return false
}
