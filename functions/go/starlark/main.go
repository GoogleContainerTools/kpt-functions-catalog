package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/starlark/starlark"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(starlark.Process)); err != nil {
		os.Exit(1)
	}
}
