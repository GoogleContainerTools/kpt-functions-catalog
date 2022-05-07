package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(setImageTags)); err != nil {
		os.Exit(1)
	}
}

func setImageTags(rl *fn.ResourceList) (bool, error) {
	si := SetImage{}
	err, ok := si.Config(rl.FunctionConfig)
	if !ok {
		return false, err
	}
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, rl.FunctionConfig))
		return true, nil
	}
	err = si.Transform(rl)
	if err != nil {
		return false, err
	}
	rl.Results = append(rl.Results, si.SdkResults()...)
	return true, nil
}
