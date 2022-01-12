package pkg

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

type ObjectScanner struct{}

// Scan searches for mutation markup comments and parses them as substitutions.
func (os *ObjectScanner) Scan(obj *yaml.RNode) (*ApplyTimeMutation, error) {
	if obj.GetKind() != "ApplyTimeMutation" {
		// no match
		return nil, nil
	}
	if obj.GetApiVersion() != "function.kpt.dev/v1alpha1" {
		// no match
		return nil, nil
	}

	config, err := obj.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format object as yaml: %w", err)
	}

	var atm ApplyTimeMutation
	err = k8syaml.Unmarshal([]byte(config), &atm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ApplyTimeMutation object: %w", err)
	}
	// TODO: validate field values (ex: non-empty)
	return &atm, nil
}
