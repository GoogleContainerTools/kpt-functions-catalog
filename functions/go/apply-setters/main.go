package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-setters/applysetters"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-setters/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

//nolint
func main() {
	asp := applysetters.ApplySettersProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.ApplySettersShort
	cmd.Long = generated.ApplySettersLong
	cmd.Example = generated.ApplySettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
