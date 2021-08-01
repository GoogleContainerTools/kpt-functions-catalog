package main

import (
	"fmt"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/contrib/go/configmap-injector/generated"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/contrib/go/configmap-injector/injector"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

//nolint
func main() {
	cip := injector.ConfigmapInjectionProcessor{}
	cmd := command.Build(&cip, command.StandaloneEnabled, false)

	cmd.Short = generated.ConfigmapInjectorShort
	cmd.Long = generated.ConfigmapInjectorLong
	cmd.Example = generated.ConfigmapInjectorExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
