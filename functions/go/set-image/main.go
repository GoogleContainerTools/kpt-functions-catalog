package main

import (
	"os"

	sdk "github.com/GoogleContainerTools/kpt-functions-catalog/thirdparty/kyaml/fnsdk"
)

func main() {
	if err := sdk.AsMain(sdk.ResourceListProcessorFunc(setImageTags)); err != nil {
		os.Exit(1)
	}
}

func setImageTags(rl *sdk.ResourceList) error {
	si := SetImage{}
	if err := si.Config(rl.FunctionConfig); err != nil {
		return err
	}
	transformedItems, err := si.Transform(rl.Items)
	if err != nil {
		return err
	}
	rl.Items = transformedItems
	return nil
}
