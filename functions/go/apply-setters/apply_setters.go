package main

import (
	"fmt"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = ApplySetters{}

// ApplySetters applies the setter values to the resource fields which are tagged
// by the setter reference comments
type ApplySetters struct {
	// Setters holds the user provided values for all the setters
	Setters []Setter `json:"setters,omitempty" yaml:"setters,omitempty"`
}

type Setter struct {
	// Name is the name of the setter
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Value is the input value for setter
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

const SetterCommentIdentifier = "# kpt-set: "

// Filter implements Set as a yaml.Filter
func (as ApplySetters) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for i := range nodes {
		err := accept(&as, nodes[i])
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return nodes, nil
}

/*
visitMapping takes input mapping node, and performs following steps
checks if the key node of the input mapping node has line comment with SetterCommentIdentifier
checks if the value node is of sequence node type
if yes to both, resolves the setter value for the setter name in the line comment
replaces the existing sequence node with the new values provided by user

e.g. for input of Mapping node

environments: # kpt-set: ${env}
- dev
- stage

For input ApplySetters [name: env, value: "[stage, prod]"], qthe yaml node is transformed to

environments: # kpt-set: ${env}
- stage
- prod

*/
func (as *ApplySetters) visitMapping(object *yaml.RNode) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		if node.IsNilOrEmpty() {
			return nil
		}
		// the aim of this method is to apply-setter for sequence nodes
		if node.Value.YNode().Kind != yaml.SequenceNode {
			// return if it is not a sequence node
			return nil
		}

		setterPattern := extractSetterPattern(node.Key)
		if setterPattern == "" {
			// the node is not tagged with setter pattern
			return nil
		}

		if !shouldSet(setterPattern, as.Setters) {
			// this means there is no intent from user to modify this setter tagged resources
			return nil
		}

		// since this setter pattern is found on sequence node, make sure that it is
		// not interpolation of setters, it should be simple setter e.g. ${environments}
		if !validArraySetterPattern(setterPattern) {
			return errors.Errorf("invalid setter pattern for array node: %q", setterPattern)
		}

		// get the setter value for the setter name in the comment
		sv := setterValue(as.Setters, setterPattern)

		// parse the setter value as yaml node
		rn, err := yaml.Parse(sv)
		if err != nil {
			return err
		}

		// the setter value must parse as sequence node
		if rn.YNode().Kind != yaml.SequenceNode {
			return errors.Errorf("input to array setter must be an array of values, but found %q", sv)
		}

		node.Value.YNode().Content = rn.YNode().Content
		node.Value.YNode().Style = yaml.FoldedStyle
		return nil
	})
}

/*
visitScalar accepts the input scalar node and performs following steps,
checks if the line comment of input scalar node has prefix SetterCommentIdentifier
resolves the setter values for the setter name in the comment
replaces the existing value of the scalar node with the new value

e.g.for input of scalar node 'nginx:1.7.1 # kpt-set: ${image}:${tag}' in the yaml node

apiVersion: v1
...
  image: nginx:1.7.1 # kpt-set: ${image}:${tag}

and for input ApplySetters [[name: image, value: ubuntu], [name: tag, value: 1.8.0]]
The yaml node is transformed to

apiVersion: v1
...
  image: ubuntu:1.8.0 # kpt-set: ${image}:${tag}

*/
func (as *ApplySetters) visitScalar(object *yaml.RNode) error {
	if object.IsNilOrEmpty() {
		return nil
	}

	if object.YNode().Kind != yaml.ScalarNode {
		// return if it is not a scalar node
		return nil
	}

	// perform a direct set of the field if it matches
	setterPattern := extractSetterPattern(object)
	if setterPattern == "" {
		// the node is not tagged with setter pattern
		return nil
	}

	curPattern := setterPattern
	if !shouldSet(setterPattern, as.Setters) {
		// this means there is no intent from user to modify this setter tagged resources
		return nil
	}

	// replace the setter names in comment pattern with provided values
	for _, setter := range as.Setters {
		setterPattern = strings.ReplaceAll(
			setterPattern,
			fmt.Sprintf("${%s}", setter.Name),
			fmt.Sprintf("%v", setter.Value),
		)
	}

	// replace the remaining setter names in comment pattern with values derived from current
	// field value, these values are not provided by user
	currentSetterValues := currentSetterValues(curPattern, object.YNode().Value)
	for setterName, setterValue := range currentSetterValues {
		setterPattern = strings.ReplaceAll(
			setterPattern,
			fmt.Sprintf("${%s}", setterName),
			fmt.Sprintf("%v", setterValue),
		)
	}

	// check if there are unresolved setters and throw error
	urs := unresolvedSetters(setterPattern)
	if len(urs) > 0 {
		return errors.Errorf("values for setters %v must be provided", urs)
	}

	object.YNode().Value = setterPattern
	object.YNode().Tag = yaml.NodeTagEmpty
	return nil
}

// shouldSet takes the setter pattern comment and setter values map and returns true
// iff at least one of the setter names in the pattern match with the setter names
// in input setterValues map
func shouldSet(pattern string, setters []Setter) bool {
	for _, s := range setters {
		if strings.Contains(pattern, fmt.Sprintf("${%s}", s.Name)) {
			return true
		}
	}
	return false
}

// currentSetterValues takes pattern and value and returns setter names to values
// derived using pattern matching
// e.g. pattern = foo-${image}:${tag}-bar, value = foo-nginx:1.7.1-bar
// returns {"image":"nginx", "tag":"1.7.1"}
func currentSetterValues(pattern, value string) map[string]string {
	res := make(map[string]string)
	// get all setter names enclosed in ${}
	urs := unresolvedSetters(pattern)
	// transform pattern replace pattern with named matching groups
	// e.g. foo-${image}:${tag}-bar => foo-(?P<image>.*):(?P<tag>.*)-bar
	for _, setterName := range urs {
		pattern = strings.ReplaceAll(
			pattern,
			setterName,
			fmt.Sprintf(`(?P<%s>.*)`, clean(setterName)))
	}
	r, err := regexp.Compile(pattern)
	if err != nil {
		// just return empty map if values can't be derived from pattern
		return res
	}
	setterValues := r.FindStringSubmatch(value)
	setterNames := r.SubexpNames()
	if len(setterNames) != len(setterValues) {
		// just return empty map if values can't be derived
		return res
	}
	for i := range setterNames {
		if i == 0 {
			// first value is just entire value, so skip it
			continue
		}
		res[setterNames[i]] = setterValues[i]
	}
	return res
}

// setterValue returns the value for the setter
func setterValue(setters []Setter, setterName string) string {
	for _, setter := range setters {
		if setter.Name == clean(setterName) {
			return setter.Value
		}
	}
	return ""
}

// extractSetterPattern extracts the setter pattern from the line comment of the
// input yaml RNode. If the the line comment doesn't contain SetterCommentIdentifier
// prefix, then it returns empty string
func extractSetterPattern(node *yaml.RNode) string {
	if node == nil {
		return ""
	}
	lineComment := node.YNode().LineComment
	if !strings.HasPrefix(lineComment, SetterCommentIdentifier) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(lineComment, SetterCommentIdentifier))
}

// validArraySetterPattern returns true if the array setter pattern is valid
// pattern must not interpolation of setters, it should be simple setter e.g. ${environments}
func validArraySetterPattern(pattern string) bool {
	return len(unresolvedSetters(pattern)) == 1 &&
		strings.HasPrefix(pattern, "${") &&
		strings.HasSuffix(pattern, "}")
}

// unresolvedSetters returns the list of values enclosed in ${} present within given
// pattern e.g. pattern = foo-${image}:${tag}-bar return ["${image}", "${tag}"]
func unresolvedSetters(pattern string) []string {
	re := regexp.MustCompile(`\$\{([^}]*)\}`)
	return re.FindAllString(pattern, -1)
}

// clean extracts value enclosed in ${}
func clean(input string) string {
	input = strings.TrimSpace(input)
	return strings.TrimSuffix(strings.TrimPrefix(input, "${"), "}")
}
