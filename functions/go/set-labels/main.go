package main

import (
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-labels/transformer"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"os"
)

//nolint
func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(transformer.SetLabels)); err != nil {
		os.Exit(1)
	}
}
