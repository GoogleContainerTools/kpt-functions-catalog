package main

import (
	"os"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/transformer"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(SetImageTagSDK)); err != nil {
		os.Exit(1)
	}
}

func SetImageTagSDK(rl *fn.ResourceList) (bool, error) {
	imageTransformer := transformer.ImageTransformer{}
	if err := imageTransformer.Config(rl.FunctionConfig); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, err
	}
	// if addtionalImageFields is supplied, run with original method
	if imageTransformer.AdditionalImageFields != nil {
		return setImageTags(rl)
	}
	if err := imageTransformer.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
		return false, err
	}

	rl.Results = append(rl.Results, imageTransformer.Results...)
	return true, nil
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
