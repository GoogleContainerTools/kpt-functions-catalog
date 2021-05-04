package applysetters

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// visitor is implemented by structs which need to walk the configuration.
// visitor is provided to accept to walk configuration
type visitor interface {
	// visitScalar is called for each scalar field value on a resource
	// node is the scalar field value
	visitScalar(node *yaml.RNode) error

	// visitMapping is called for each Mapping field value on a resource
	// node is the mapping field value
	visitMapping(node *yaml.RNode) error
}

// accept invokes the appropriate function on v for each field in object
func accept(v visitor, object *yaml.RNode) error {
	return acceptImpl(v, object)
}

// acceptImpl implements accept using recursion
func acceptImpl(v visitor, object *yaml.RNode) error {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		// Traverse the child of the document
		return accept(v, object)
	case yaml.MappingNode:
		if err := v.visitMapping(object); err != nil {
			return err
		}
		return object.VisitFields(func(node *yaml.MapNode) error {
			// Traverse each field value
			return acceptImpl(v, node.Value)
		})
	case yaml.SequenceNode:
		return object.VisitElements(func(node *yaml.RNode) error {
			// Traverse each list element
			return acceptImpl(v, node)
		})
	case yaml.ScalarNode:
		// Visit the scalar field
		return v.visitScalar(object)
	}
	return nil
}
