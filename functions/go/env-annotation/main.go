// Package main implements a templater kpt-function, that
// allows to generate documents using go-tempates.
// Templater allows to use Sprig functions, including env and expandenv[1].
// All other fields except `template` are passed as values to go-template
// engine.
// [1] http://masterminds.github.io/sprig/os.html

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/env-annotation/function"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

func main() {
	cfg := function.Config{}
	resourceList := &framework.ResourceList{FunctionConfig: &cfg}

	cmd := framework.Command(resourceList, func() error {
		fn, err := function.NewFilter(&cfg)
		if err != nil {
			log.Printf("function creation failed: %v", err)
			return err
		}

		items, err := fn.Filter(resourceList.Items)
		if err != nil {
			return err
		}
		resourceList.Items = items
		return nil
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
