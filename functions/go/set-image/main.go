package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/transformer"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(&transformer.SetImage{}); err != nil {
		os.Exit(1)
	}
}
