package createsetters

import (
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = &CreateSetters{}

// CreateSetters creates a comment for the resource fields which
// contain the same value as setter value
type CreateSetters struct {
	// ScalarSetters holds the user provided values for simple scalar setters
	ScalarSetters []ScalarSetter

	// ArraySetters holds the user provided values for array setters
	ArraySetters []ArraySetter

	// Results are the results of adding setter comments
	Results []*Result

	// filePath file path of resource
	filePath string
}

// ScalarSetter stores name and value of the map setter
type ScalarSetter struct {
	// Name is the name of the setter
	Name string

	// Value is the value of the field to which setter comment is added.
	Value string
}

// ArraySetter stores name and values of the array setter
type ArraySetter struct {
	// Name is the name of the setter
	Name string

	// Values are the values of the field to which setter comment is added.
	Values []string
}

// Result holds result of create-setters operation
type Result struct {
	// FilePath is the file path of the matching value
	FilePath string

	// FieldPath is field path of the matching value
	FieldPath string

	// Value is the value of the field to which setter comment is added.
	Value string

	// Comment is the line comment of the matching value
	Comment string
}

// CompareSetters is to sort the setter values
type CompareSetters []ScalarSetter

func (a CompareSetters) Len() int {
	return len(a)
}

// checks if the jth node's Name is a substring of ith node's Value
// jth node is sorted first if the condition statisfies
func (a CompareSetters) Less(i, j int) bool {
	return !strings.Contains(a[i].Value, a[j].Name)
}

func (a CompareSetters) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Filter implements CreatSetters cs a yaml.Filter
func (cs *CreateSetters) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {

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
checks if all the values in the node match any of the ArraySetters
adds the linecomment if they are equal

checks if any of the values of node matches with ScalarSetters
changes the node to FoldedStyle

e.g. for input of Mapping node

environments:
  - dev
  - stage

For input CreateSetters [Name: env, Values: [dev, stage]], yaml node is transformed to

environments: # kpt-set: ${env}
  - dev
  - stage

e.g. for input of Mapping node with FlowStyle

env: [foo, bar]

For input CreateSetters [Name: env, Values: [foo, bar]], yaml node is transformed to

env: [foo, bar] # kpt-set: ${env}

e.g. for input of Mapping node with FlowStyle matching few values from ScalarSetters

env: [foo, bar]

For input CreateSetters [Name: image, Values: foo], yaml node is transformed to

env:
  - foo
  - bar
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
		sort.Strings(nodeValues)

		// checks if any of the values of node matches with ScalarSetters
		// changes the node to FoldedStyle
		nodeToAddComment := node.Value
		if nodeToAddComment.YNode().Style == yaml.FlowStyle {
			if hasMatchValue(nodeValues, cs.ScalarSetters) {
				// changing the node style to FoldedStyle
				nodeToAddComment.YNode().Style = yaml.FoldedStyle
				// To add the comment to the key for the FoldedStyle value node
				nodeToAddComment = node.Key
			}
		}

		for _, arraySetters := range cs.ArraySetters {
			// checks if all the values in node are present in array setter
			if checkEqual(nodeValues, arraySetters.Values) {
				if nodeToAddComment.YNode().Style == yaml.FlowStyle && len(nodeValues) > 0 {
					nodeToAddComment.YNode().Style = yaml.FoldedStyle
					nodeToAddComment = node.Key
				}
				nodeToAddComment.YNode().LineComment = fmt.Sprintf("kpt-set: ${%s}", arraySetters.Name)
				cs.Results = append(cs.Results, &Result{
					FilePath:  cs.filePath,
					FieldPath: fieldPath,
					Value:     fmt.Sprint(nodeValues),
					Comment:   nodeToAddComment.YNode().LineComment,
				})
				return nil
			}
		}
		return nil
	})
}

/**
visitScalar accepts the input scalar node and performs following steps,
checks if it is a scalar node
adds the linecomment if it's value matches with any of the setter

e.g.for input of scalar node 'nginx:1.7.1' in the yaml node

apiVersion: v1
...
  env: dev
  image: nginx:1.7.1

and for input CreateSetters [[name: image, value: nginx], [name: env, value: dev], [name: tag, value: 1.7.1]]
The yaml node is transformed to

apiVersion: v1
...
  env: dev # kpt-set: ${env}
  image: nginx:1.7.1 # kpt-set: ${image}:${tag}

*/

func (cs *CreateSetters) visitScalar(object *yaml.RNode, path string) error {
	if object.YNode().Kind != yaml.ScalarNode {
		// return if it is not a scalar node
		return nil
	}

	linecomment, valueMatch := getLineComment(object.YNode().Value, cs.ScalarSetters)

	// sets the linecomment if the match is found
	if valueMatch {
		object.YNode().LineComment = fmt.Sprintf("kpt-set: %s", linecomment)
		cs.Results = append(cs.Results, &Result{
			FilePath:  cs.filePath,
			FieldPath: strings.TrimPrefix(path, "."),
			Value:     object.YNode().Value,
			Comment:   object.YNode().LineComment,
		})
	}

	return nil
}

/**
Decode decodes the input yaml node into CreatSetters struct
places the setter either in ScalarSetters or ArraySetters
sorts the ScalarSetters using CompareSetters

e.g.for input ScalarSetters
	[[name: image, value: nginx], [name: ubuntu, value: image]]

Input:
	spec: nginx-development

To avoid this case, sorts the ScalarSetters
	spec: nginx-development # kpt-set: ${ubuntu}-development

ScalarSetters array is transformed to
	[[name: ubuntu, value: image], [name: image, value: nginx]]
*/
func Decode(rn *yaml.RNode, fcd *CreateSetters) error {
	if len(rn.GetDataMap()) == 0 {
		return fmt.Errorf("config map cannot be empty")
	}
	for k, v := range rn.GetDataMap() {
		parsedInput, err := yaml.Parse(v)
		if err != nil {
			return fmt.Errorf("parsing error")
		}
		// checks if the value is SequenceNode
		// adds to the ArraySetters if it is a SequenceNode
		// adds to the ScalarSetters if it is a ScalarNode
		if parsedInput.YNode().Kind == yaml.SequenceNode {
			fcd.ArraySetters = append(fcd.ArraySetters, ArraySetter{Name: k, Values: getArraySetter(parsedInput)})
		} else if parsedInput.YNode().Kind == yaml.ScalarNode {
			fcd.ScalarSetters = append(fcd.ScalarSetters, ScalarSetter{Name: k, Value: v})
		}
	}

	// sorts all the Setters
	sort.Sort(CompareSetters(fcd.ScalarSetters))
	return nil
}

// checkEqual checks if all the values in node are present in array setter
func checkEqual(nodeValues []string, arraySetters []string) bool {
	if len(nodeValues) != len(arraySetters) {
		return false
	}

	for idx := range nodeValues {
		if arraySetters[idx] != nodeValues[idx] {
			return false
		}
	}
	return true
}

// getArraySetter parses the input and returns array setters
func getArraySetter(input *yaml.RNode) []string {
	output := []string{}

	elements, err := input.Elements()
	if err != nil {
		return output
	}

	for _, as := range elements {
		output = append(output, as.YNode().Value)
	}

	sort.Strings(output)
	return output
}

// hasMatchValue checks if any of the ScalarSetter value matches with the node value
func hasMatchValue(nodeValues []string, setters []ScalarSetter) bool {
	for _, value := range nodeValues {
		for _, setter := range setters {
			if strings.Contains(value, setter.Value) {
				return true
			}
		}
	}
	return false
}
/**
getLineComment checks if any of the setters value matches with the node value
replaces that part of the node value with the ${setterName}
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
func getLineComment(nodeValue string, setters []ScalarSetter) (string, bool) {
	output := nodeValue
	valueMatch := false

	for _, setter := range setters {
		if strings.Contains(nodeValue, setter.Value) {
			valueMatch = true
			output = strings.ReplaceAll(
				output,
				setter.Value,
				fmt.Sprintf("${%s}", setter.Name),
			)
		}
	}

	return output, valueMatch
}
