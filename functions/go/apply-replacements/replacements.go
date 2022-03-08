package main

import (
	"fmt"
	"io"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
	"sigs.k8s.io/kustomize/api/filters/replacement"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

type Replacements struct {
	Replacements []types.Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

const fnConfigKind = "Replacements"

// Config initializes Replacements from a functionConfig sdk.KubeObject
func (r *Replacements) Config(functionConfig *sdk.KubeObject) error {
	if functionConfig.Kind() != fnConfigKind {
		return fmt.Errorf("received functionConfig of kind %s, only functionConfig of kind %s is supported",
			functionConfig.Kind(), fnConfigKind)
	}
	r.Replacements = []types.Replacement{}
	if err := functionConfig.As(r); err != nil {
		return fmt.Errorf("unable to convert functionConfig to %s:\n%w",
			"replacements", err)
	}
	return nil
}

// Run accepts a reader and writer, and calls framework.Execute to run the function. This exists
// to enable the function to be invoked as a library.
func (r *Replacements) Run(reader io.Reader, writer io.Writer) error {
	rw := &kio.ByteReadWriter{
		Reader:                reader,
		Writer:                writer,
		KeepReaderAnnotations: true,
	}
	return framework.Execute(r, rw)
}

// Process implements framework.ResourceListProcessor interface.
func (r *Replacements) Process(rl *framework.ResourceList) error {
	if err := r.Config(sdk.NewFromRNode(rl.FunctionConfig)); err != nil {
		return err
	}
	transformedItems, err := replacement.Filter{Replacements: r.Replacements}.Filter(rl.Items)
	if err != nil {
		return err
	}
	rl.Items = transformedItems
	return nil
}
