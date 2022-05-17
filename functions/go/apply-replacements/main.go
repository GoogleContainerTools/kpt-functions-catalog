package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/apply-replacements/replacements"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(replacements.ApplyReplacements)); err != nil {
		os.Exit(1)
	}
}
